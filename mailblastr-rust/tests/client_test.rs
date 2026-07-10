//! Integration tests that drive the real client against a tiny in-process
//! HTTP/1.1 stub (std `TcpListener`, one request per test) via the
//! `Mailblastr::with_base_url` override — no network, no mocks framework.

use std::io::{Read, Write};
use std::net::TcpListener;
use std::thread::{self, JoinHandle};

use mailblastr::{CreateEmailBaseOptions, Error, ListSegmentsParams, Mailblastr, PaginationParams};

/// Find the first occurrence of `needle` in `haystack`.
fn find_subsequence(haystack: &[u8], needle: &[u8]) -> Option<usize> {
    haystack
        .windows(needle.len())
        .position(|window| window == needle)
}

/// Spawn a one-shot HTTP stub. Returns the base URL and a handle that yields
/// the raw request (start-line + headers + body) once the stub has answered.
fn spawn_stub(status: u16, body: &'static str) -> (String, JoinHandle<String>) {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind stub listener");
    let addr = listener.local_addr().expect("stub addr");
    let handle = thread::spawn(move || {
        let (mut stream, _) = listener.accept().expect("accept");
        let mut buf: Vec<u8> = Vec::new();
        let mut tmp = [0u8; 4096];
        loop {
            let n = stream.read(&mut tmp).expect("read request");
            if n == 0 {
                break;
            }
            buf.extend_from_slice(&tmp[..n]);
            if let Some(header_end) = find_subsequence(&buf, b"\r\n\r\n") {
                let headers = String::from_utf8_lossy(&buf[..header_end]).to_string();
                let content_length = headers
                    .lines()
                    .find_map(|line| {
                        let (name, value) = line.split_once(':')?;
                        if name.trim().eq_ignore_ascii_case("content-length") {
                            value.trim().parse::<usize>().ok()
                        } else {
                            None
                        }
                    })
                    .unwrap_or(0);
                if buf.len() >= header_end + 4 + content_length {
                    break;
                }
            }
        }
        let reason = match status {
            200 => "OK",
            404 => "Not Found",
            422 => "Unprocessable Entity",
            _ => "Status",
        };
        let response = format!(
            "HTTP/1.1 {status} {reason}\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{body}",
            body.len(),
        );
        stream
            .write_all(response.as_bytes())
            .expect("write response");
        let _ = stream.flush();
        String::from_utf8_lossy(&buf).to_string()
    });
    (format!("http://{addr}"), handle)
}

/// Spawn a stub that answers a SEQUENCE of responses, one per incoming
/// connection (the client sends `Connection: close`, so each retry reconnects).
/// The thread is detached: a broken retry loop makes fewer requests, which
/// surfaces as the client call failing — never as a hang.
fn spawn_stub_seq(responses: Vec<(u16, &'static str, Option<&'static str>)>) -> String {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind stub listener");
    let addr = listener.local_addr().expect("stub addr");
    thread::spawn(move || {
        for (status, body, retry_after) in responses {
            let (mut stream, _) = match listener.accept() {
                Ok(s) => s,
                Err(_) => break,
            };
            let mut buf: Vec<u8> = Vec::new();
            let mut tmp = [0u8; 4096];
            loop {
                let n = match stream.read(&mut tmp) {
                    Ok(n) => n,
                    Err(_) => break,
                };
                if n == 0 {
                    break;
                }
                buf.extend_from_slice(&tmp[..n]);
                if let Some(header_end) = find_subsequence(&buf, b"\r\n\r\n") {
                    let headers = String::from_utf8_lossy(&buf[..header_end]);
                    let content_length = headers
                        .lines()
                        .find_map(|line| {
                            let (name, value) = line.split_once(':')?;
                            if name.trim().eq_ignore_ascii_case("content-length") {
                                value.trim().parse::<usize>().ok()
                            } else {
                                None
                            }
                        })
                        .unwrap_or(0);
                    if buf.len() >= header_end + 4 + content_length {
                        break;
                    }
                }
            }
            let ra = retry_after
                .map(|v| format!("Retry-After: {v}\r\n"))
                .unwrap_or_default();
            let response = format!(
                "HTTP/1.1 {status} Status\r\nContent-Type: application/json\r\n{ra}Content-Length: {}\r\nConnection: close\r\n\r\n{body}",
                body.len(),
            );
            let _ = stream.write_all(response.as_bytes());
            let _ = stream.flush();
        }
    });
    format!("http://{addr}")
}

#[tokio::test]
async fn retries_a_429_then_succeeds() {
    // First attempt is rate-limited (Retry-After: 0 → no real wait); the retry
    // gets a 200. Reaching "em_ok" proves the client reconnected and retried.
    let base_url = spawn_stub_seq(vec![
        (
            429,
            r#"{"statusCode":429,"name":"rate_limited","message":"slow down"}"#,
            Some("0"),
        ),
        (200, r#"{"id":"em_ok"}"#, None),
    ]);
    let mb = Mailblastr::with_base_url("mb_test_key", base_url);
    let email = CreateEmailBaseOptions::new("Acme <hi@x.com>", ["b@y.com"], "Hi").with_html("<p>x</p>");
    let sent = mb
        .emails
        .send(email)
        .await
        .expect("send should succeed after retrying the 429");
    assert_eq!(sent.id, "em_ok");
}

#[tokio::test]
async fn does_not_retry_a_422() {
    // A single 422 stub: if the client wrongly retried, the second attempt would
    // hit a closed listener and the error would change — it must surface the 422.
    let base_url = spawn_stub_seq(vec![(
        422,
        r#"{"statusCode":422,"name":"validation_error","message":"bad"}"#,
        None,
    )]);
    let mb = Mailblastr::with_base_url("mb_test_key", base_url);
    let email = CreateEmailBaseOptions::new("Acme <hi@x.com>", ["b@y.com"], "Hi").with_html("<p>x</p>");
    let err = mb.emails.send(email).await.expect_err("422 should fail");
    match err {
        Error::Api { status_code, .. } => assert_eq!(status_code, 422),
        other => panic!("expected Error::Api(422), got {other:?}"),
    }
}

#[tokio::test]
async fn sends_an_email_with_auth_and_json_body() {
    let (base_url, handle) = spawn_stub(200, r#"{"id":"em_123"}"#);
    let mb = Mailblastr::with_base_url("mb_test_key", base_url);

    let email =
        CreateEmailBaseOptions::new("Acme <hello@yourdomain.com>", ["user@example.com"], "Hello")
            .with_html("<p>Hi</p>");

    let sent = mb.emails.send(email).await.expect("send should succeed");
    assert_eq!(sent.id, "em_123");

    let request = handle.join().unwrap();
    let lower = request.to_lowercase();
    assert!(
        request.starts_with("POST /emails HTTP/1.1"),
        "got: {request}"
    );
    assert!(lower.contains("authorization: bearer mb_test_key"));
    assert!(lower.contains("user-agent: mailblastr-rust/"));
    assert!(request.contains(r#""from":"Acme <hello@yourdomain.com>""#));
    assert!(request.contains(r#""to":["user@example.com"]"#));
    assert!(request.contains(r#""html":"<p>Hi</p>""#));
    // Unset optional fields must be omitted, not sent as null.
    assert!(!request.contains(r#""text""#));
}

#[tokio::test]
async fn maps_non_2xx_to_api_error() {
    let (base_url, handle) = spawn_stub(
        422,
        r#"{"statusCode":422,"name":"validation_error","message":"domain is required"}"#,
    );
    let mb = Mailblastr::with_base_url("mb_test_key", base_url);

    let err = mb.domains.get("dom_1").await.expect_err("should fail");
    match err {
        Error::Api {
            status_code,
            name,
            message,
        } => {
            assert_eq!(status_code, 422);
            assert_eq!(name, "validation_error");
            assert_eq!(message, "domain is required");
        }
        other => panic!("expected Error::Api, got {other:?}"),
    }
    handle.join().unwrap();
}

#[tokio::test]
async fn segments_list_requires_domain_in_query() {
    let (base_url, handle) = spawn_stub(200, r#"{"object":"list","has_more":false,"data":[]}"#);
    let mb = Mailblastr::with_base_url("mb_test_key", base_url);

    let page = mb
        .segments
        .list(ListSegmentsParams::new("example.com").with_limit(10))
        .await
        .expect("list should succeed");
    assert!(page.data.is_empty());
    assert!(!page.has_more);

    let request = handle.join().unwrap();
    assert!(
        request.starts_with("GET /segments?domain=example.com&limit=10 HTTP/1.1"),
        "got: {request}"
    );
}

#[tokio::test]
async fn contact_email_ids_are_percent_encoded_in_the_path() {
    let (base_url, handle) = spawn_stub(
        200,
        r#"{"object":"contact","id":"con_1","email":"user@example.com","first_name":null,"last_name":null,"unsubscribed":false,"created_at":"2026-01-01T00:00:00Z"}"#,
    );
    let mb = Mailblastr::with_base_url("mb_test_key", base_url);

    let contact = mb
        .contacts
        .get(mailblastr::ContactLookup::new("user@example.com").with_domain("example.com"))
        .await
        .expect("get should succeed");
    assert_eq!(contact.id, "con_1");

    let request = handle.join().unwrap();
    assert!(
        request.starts_with("GET /contacts/user%40example.com?domain=example.com HTTP/1.1"),
        "got: {request}"
    );
}

#[tokio::test]
async fn idempotency_key_header_is_sent() {
    let (base_url, handle) = spawn_stub(200, r#"{"id":"em_9"}"#);
    let mb = Mailblastr::with_base_url("mb_test_key", base_url);

    let email = CreateEmailBaseOptions::new("a@b.com", ["c@d.com"], "S");
    mb.emails
        .send_with_idempotency_key(email, "order-123")
        .await
        .expect("send should succeed");

    let request = handle.join().unwrap();
    assert!(request
        .to_lowercase()
        .contains("idempotency-key: order-123"));
}

#[tokio::test]
async fn received_attachment_downloads_return_raw_bytes() {
    let (base_url, handle) = spawn_stub(200, "%PDF-1.7 raw bytes, not json");
    let mb = Mailblastr::with_base_url("mb_test_key", base_url);

    let bytes = mb
        .emails
        .receiving
        .get_attachment("re_1", "att_1")
        .await
        .expect("download should succeed");
    assert_eq!(bytes, b"%PDF-1.7 raw bytes, not json".to_vec());

    let request = handle.join().unwrap();
    assert!(
        request.starts_with("GET /emails/receiving/re_1/attachments/att_1 HTTP/1.1"),
        "got: {request}"
    );
}

#[tokio::test]
async fn pagination_params_apply_to_list_endpoints() {
    let (base_url, handle) = spawn_stub(200, r#"{"object":"list","has_more":true,"data":[]}"#);
    let mb = Mailblastr::with_base_url("mb_test_key", base_url);

    let page = mb
        .emails
        .list(Some(
            PaginationParams::new().with_limit(25).with_after("em_5"),
        ))
        .await
        .expect("list should succeed");
    assert!(page.has_more);

    let request = handle.join().unwrap();
    assert!(
        request.starts_with("GET /emails?limit=25&after=em_5 HTTP/1.1"),
        "got: {request}"
    );
}
