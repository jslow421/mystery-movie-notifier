use lambda_runtime::{Error, LambdaEvent, service_fn};
use rquest::Client;
use rquest_util::Emulation;
use serde_json::{Value, json};

#[tokio::main]
async fn main() -> Result<(), Error> {
    // 1. Logic that only runs once during "Cold Start" goes here (e.g., client init)

    // 2. Register the handler
    let func = service_fn(handler);
    lambda_runtime::run(func).await?;
    Ok(())
}

async fn handler(event: LambdaEvent<Value>) -> Result<Value, Error> {
    // 3. Your scraping logic or function calls go here
    let payload = event.payload;

    Ok(json!({ "message": "Scraper executed successfully"}))
}

async fn run_scrape() -> Result<(), rquest::Error> {
    // Build client
    let client = Client::builder().emulation(Emulation::Firefox136).build()?;

    // Use the API you're familiar with
    let resp = client
        .get("https://www.marcustheatres.com/marcus-specials/marcus-film-series/marcus-mystery-movie")
        .send()
        .await?;
    println!("{}", resp.text().await?);

    Ok(())
}

#[cfg(test)]
mod test {
    use super::*;

    #[tokio::test]
    async fn test_run_scrape() {
        let result = run_scrape().await;
        assert!(result.is_ok());
    }
}
