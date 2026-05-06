import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";

export class InfraStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);
    // Create S3 bucket to hold the compiled functions for lambdas
    const bucket = new cdk.aws_s3.Bucket(this, "MovieNotifierBucket", {
      versioned: false,
      removalPolicy: cdk.RemovalPolicy.DESTROY, // Only for dev/test environments
      autoDeleteObjects: true, // Only for dev/test environments
    });

    // Lambda function to scrape the website and get the data from the element
    const scraperFunction = new cdk.aws_lambda.Function(this, "ScraperFunction", {
      runtime: cdk.aws_lambda.Runtime.PROVIDED_AL2023,
      code: cdk.aws_lambda.Code.fromBucket(bucket, "retriever"), // Specific name for retriever binary
      handler: "index.handler", // This is ignored for custom runtime
      environment: {
        BUCKET_NAME: bucket.bucketName,
      },
    });
  }
}
