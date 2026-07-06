package com.mailblastr.http;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.Map;

/** Default {@link HttpTransport} backed by {@code java.net.http.HttpClient} (Java 11+). */
public final class DefaultHttpTransport implements HttpTransport {
    private final HttpClient client;
    private final Duration requestTimeout;

    public DefaultHttpTransport() {
        this(HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(30)).build(),
                Duration.ofSeconds(60));
    }

    public DefaultHttpTransport(HttpClient client, Duration requestTimeout) {
        this.client = client;
        this.requestTimeout = requestTimeout;
    }

    @Override
    public HttpResult execute(String method, String url, Map<String, String> headers, String body) throws Exception {
        HttpRequest.Builder builder = HttpRequest.newBuilder()
                .uri(URI.create(url))
                .timeout(requestTimeout)
                .method(method, body == null
                        ? HttpRequest.BodyPublishers.noBody()
                        : HttpRequest.BodyPublishers.ofString(body));
        for (Map.Entry<String, String> e : headers.entrySet()) {
            builder.header(e.getKey(), e.getValue());
        }
        HttpResponse<byte[]> res = client.send(builder.build(), HttpResponse.BodyHandlers.ofByteArray());
        return new HttpResult(res.statusCode(), res.body());
    }
}
