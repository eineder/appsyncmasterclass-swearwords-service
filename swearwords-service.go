package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/aws-cdk-go/awscdk/v2/triggers"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const SWEARWORDS_TABLE_NAME string = "SWEARWORDS_TABLE_NAME"
const BUCKET_NAME string = "BUCKET_NAME"
const BUCKET_KEY string = "BUCKET_KEY"

type SwearwordsServiceStackProps struct {
	awscdk.StackProps
}

func NewSwearwordsServiceStack(scope constructs.Construct, id string, props *SwearwordsServiceStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	bucket := awss3.NewBucket(stack, jsii.String("swearwords-bucket"), &awss3.BucketProps{})
	cdkAsset := awss3deployment.Source_Asset(jsii.String("assets"), &awss3assets.AssetOptions{})
	awss3deployment.NewBucketDeployment(
		stack,
		jsii.String("swearwords-bucket-deployment"),
		&awss3deployment.BucketDeploymentProps{
			Sources:           &[]awss3deployment.ISource{cdkAsset},
			DestinationBucket: bucket,
		})

	table := awsdynamodb.NewTable(stack, jsii.String("swearwords-table"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("word"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode: awsdynamodb.BillingMode_PAY_PER_REQUEST,
	})

	fn := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("is-swearword-lambda"), &awscdklambdagoalpha.GoFunctionProps{
		Entry:   jsii.String("is_swearword/is_swearword.go"),
		Runtime: awslambda.Runtime_PROVIDED_AL2(),
		Environment: &map[string]*string{
			SWEARWORDS_TABLE_NAME: table.TableName()},
	})
	fn.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"dynamodb:GetItem"),
		Resources: &[]*string{table.TableArn()},
	}))

	triggerFunc := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("prime-swearwords-lambda"), &awscdklambdagoalpha.GoFunctionProps{
		Entry:   jsii.String("prime_swearwords/prime_swearwords.go"),
		Runtime: awslambda.Runtime_PROVIDED_AL2(),
		Environment: &map[string]*string{
			SWEARWORDS_TABLE_NAME: table.TableName(),
			BUCKET_NAME:           bucket.BucketName(),
			BUCKET_KEY:            jsii.String("swearwords.txt")},
	})
	triggerFunc.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"dynamodb:GetItem",
			"dynamodb:ImportTable",
			"dynamodb:DescribeImport",
			"dynamodb:ListImports"),
		Resources: &[]*string{table.TableArn()},
	}))
	triggerFunc.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings("s3:GetObject",
			"s3:ListBucket"),
		Resources: jsii.Strings(*bucket.BucketArn(), *bucket.BucketArn()+"/*"),
	}))
	triggerFunc.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings("logs:CreateLogGroup",
			"logs:CreateLogStream",
			"logs:DescribeLogGroups",
			"logs:DescribeLogStreams",
			"logs:PutLogEvents",
			"logs:PutRetentionPolicy"),
		Resources: jsii.Strings("*"),
	}))
	triggers.NewTrigger(stack, jsii.String("prime-swearwords-trigger"), &triggers.TriggerProps{
		Handler:        triggerFunc,
		InvocationType: triggers.InvocationType_EVENT,
	})

	awscdk.NewCfnOutput(stack, jsii.String("BucketNameOutput"), &awscdk.CfnOutputProps{
		Key:   jsii.String("BucketName"),
		Value: bucket.BucketName(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("BucketKeyOutput"), &awscdk.CfnOutputProps{
		Key:   jsii.String("BucketKey"),
		Value: jsii.String("swearwords.txt"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("SwearwordsTableNameOutput"), &awscdk.CfnOutputProps{
		Key:   jsii.String("SwearwordsTableName"),
		Value: table.TableName(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewSwearwordsServiceStack(app, "SwearwordsServiceStack", &SwearwordsServiceStackProps{
		awscdk.StackProps{},
	})

	app.Synth(nil)
}
