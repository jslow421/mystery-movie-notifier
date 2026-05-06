use lambda_runtime::{Error, LambdaEvent, service_fn};
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
    let _payload = event.payload;

    Ok(json!({ "message": "Scraper executed successfully" }))
}
