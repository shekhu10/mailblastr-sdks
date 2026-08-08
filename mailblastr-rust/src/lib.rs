//! Official Rust SDK for the [MailBlastr](https://www.mailblastr.com) email API.
//!
//! ```no_run
//! use mailblastr::{CreateEmailBaseOptions, Mailblastr, Result};
//!
//! #[tokio::main]
//! async fn main() -> Result<()> {
//!     let mailblastr = Mailblastr::new("mb_xxxxxxxxx");
//!
//!     let email = CreateEmailBaseOptions::new(
//!         "Acme <hello@yourdomain.com>",
//!         ["user@example.com"],
//!         "Hello from MailBlastr",
//!     )
//!     .with_html("<p>Your first email 🎉</p>");
//!
//!     let sent = mailblastr.emails.send(email).await?;
//!     println!("sent {}", sent.id);
//!     Ok(())
//! }
//! ```
//!
//! Every service method is `async` and returns [`Result<T>`](Result), where the
//! error is [`Error`] — API failures surface as [`Error::Api`] carrying an
//! [`ApiError`], parsed from the standard MailBlastr
//! `{ statusCode, name, message }` body.
//!
//! Some errors add fields on top of that envelope, and [`ApiError`] models each
//! one as an `Option` (or an empty `Vec`) that is absent on ordinary errors:
//! [`limit`](ApiError::limit) says WHICH plan or quota allowance ran out,
//! [`reputation`](ApiError::reputation) what a reputation gate paused, and
//! [`sent`](ApiError::sent) / [`sent_count`](ApiError::sent_count) which emails
//! already went out when a batch failed part way through.
//!
//! ```no_run
//! # async fn run(mb: mailblastr::Mailblastr, email: mailblastr::CreateEmailBaseOptions) {
//! match mb.emails.send(email).await {
//!     Ok(sent) => println!("sent {}", sent.id),
//!     Err(err) => match err.api() {
//!         Some(api) => match &api.limit {
//!             Some(limit) => println!("{} cap hit: {}/{}", limit.kind, limit.used, limit.limit),
//!             None => println!("{} ({})", api.message, api.name),
//!         },
//!         None => println!("transport failure: {err}"),
//!     },
//! }
//! # }
//! ```
//!
//! # Domain-first parameters
//!
//! MailBlastr is *domain-first*: each verified sending domain owns its own
//! contact pool. `domain` is REQUIRED when creating contacts through the flat
//! `/contacts` API, on segments, topics and campaigns, on
//! [`automations.create`](services::AutomationsSvc::create), and on
//! [`events.send`](services::EventsSvc::send). See the README for details.

mod client;
pub mod services;
pub mod types;

pub use client::{
    Mailblastr, MailblastrBuilder, DEFAULT_BASE_URL, DEFAULT_MAX_RETRIES, DEFAULT_TIMEOUT_SECS,
    IDEMPOTENCY_KEY_MAX_LEN, USER_AGENT, VERSION,
};
pub use services::*;
pub use types::*;
