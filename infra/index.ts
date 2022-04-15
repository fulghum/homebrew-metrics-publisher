import events = require('aws-cdk-lib/aws-events');
import targets = require('aws-cdk-lib/aws-events-targets');
import lambda_go = require('@aws-cdk/aws-lambda-go-alpha');
import ssm = require('aws-cdk-lib/aws-ssm');
import cdk = require('aws-cdk-lib');

export class HomebrewMetricsPublisherStack extends cdk.Stack {
  constructor(app: cdk.App, id: string) {
    super(app, id)

    const lambdaGoFn = new lambda_go.GoFunction(this, 'HomebrewStatsUploader', {
      entry: "../go/cmd/homebrew-metric-publisher",
    })

    const authTokenSecureValue = ssm.StringParameter.fromSecureStringParameterAttributes(this, 'AuthTokenSecureValue', {
      parameterName: 'dolthub-auth-token',
    });
    authTokenSecureValue.grantRead(lambdaGoFn)

    // Run at 10 minutes past every other hour
    // See https://docs.aws.amazon.com/lambda/latest/dg/tutorial-scheduled-events-schedule-expressions.html
    const rule = new events.Rule(this, 'Rule', {
      schedule: events.Schedule.expression('cron(/10 */2 ? * * *)')
    })

    rule.addTarget(new targets.LambdaFunction(lambdaGoFn))
  }
}

const app = new cdk.App();
new HomebrewMetricsPublisherStack(app, 'HomebrewMetricsPublisherStack');
app.synth();
