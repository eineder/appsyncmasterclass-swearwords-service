const fs = require("fs");
const cf = require("@aws-sdk/client-cloudformation");

const cmd = new cf.DescribeStacksCommand({
  StackName: "AwsFileToDynamodbStack",
});

const client = new cf.CloudFormationClient();
client.send(cmd).then((data) => {
  const outputs = data.Stacks[0].Outputs;
  let envContent = "";
  outputs.forEach((output) => {
    envContent += `${camelToSnakeCase(output.OutputKey)}=${
      output.OutputValue
    }\n`;
  });
  fs.writeFileSync(".env", envContent);
});

function camelToSnakeCase(str) {
  return str
    .replace(/[A-Z]/g, (letter, index) => `${index > 0 ? "_" : ""}${letter}`)
    .toUpperCase();
}
