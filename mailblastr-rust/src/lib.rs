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
//! error is [`Error`] — API failures surface as
//! [`Error::Api { status_code, name, message }`](Error::Api), parsed from the
//! standard MailBlastr error body.
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
    USER_AGENT, VERSION,
};
pub use services::*;
pub use types::*;
