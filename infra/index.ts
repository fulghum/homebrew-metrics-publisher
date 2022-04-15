import cdk = require('aws-cdk-lib');
import {HomebrewMetricsPublisher} from "./lib/homebrew-publisher";

export class HomebrewMetricsPublisherStack extends cdk.Stack {
  constructor(app: cdk.App, id: string) {
    super(app, id)

    new HomebrewMetricsPublisher(this, "HomebrewMetricsUploader", {
      homebrewFormula: "dolt",
      dolthubAuthTokenParameterName: "dolthub-auth-token",
    })
  }
}

const app = new cdk.App();
new HomebrewMetricsPublisherStack(app, 'HomebrewMetricsPublisherStack');
app.synth();
