import events = require('aws-cdk-lib/aws-events');
import targets = require('aws-cdk-lib/aws-events-targets');
import lambda_go = require('@aws-cdk/aws-lambda-go-alpha');
import ssm = require('aws-cdk-lib/aws-ssm');
import {Construct} from "constructs";

export interface HomebrewMetricsPublisherProps {
    homebrewFormula: string;
    dolthubAuthTokenParameterName: string;
}

export class HomebrewMetricsPublisher extends Construct {
    constructor(scope: Construct, id: string, props: HomebrewMetricsPublisherProps) {
        super(scope, id);

        const lambdaGoFn = new lambda_go.GoFunction(this, 'HomebrewStatsUploader', {
            entry: "../go/cmd/homebrew-metric-publisher",
            environment: {
                "homebrewFormula": props.homebrewFormula,
                "dolthubAuthTokenParameterName": props.dolthubAuthTokenParameterName,
            }
        })

        const authTokenSecureValue = ssm.StringParameter.fromSecureStringParameterAttributes(this, 'AuthTokenSecureValue', {
            parameterName: props.dolthubAuthTokenParameterName,
        });
        authTokenSecureValue.grantRead(lambdaGoFn)

        // Run at 10 minutes past every four hours
        // See https://docs.aws.amazon.com/lambda/latest/dg/tutorial-scheduled-events-schedule-expressions.html
        const rule = new events.Rule(this, 'SchedulingRule', {
            schedule: events.Schedule.expression('cron(/10 */4 ? * * *)')
        })

        rule.addTarget(new targets.LambdaFunction(lambdaGoFn))
    }
}