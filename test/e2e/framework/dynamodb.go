/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package framework

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	. "github.com/onsi/gomega"
)

func (f *Framework) DynamoDBCreateTable(region, table string, readCapacity, writeCapacity int) error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	db := dynamodb.New(sess, aws.NewConfig().WithRegion(region))

	in := &dynamodb.CreateTableInput{
		TableName: aws.String(table),
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(int64(readCapacity)),
			WriteCapacityUnits: aws.Int64(int64(writeCapacity)),
		},
		KeySchema: []*dynamodb.KeySchemaElement{{
			AttributeName: aws.String("Path"),
			KeyType:       aws.String("HASH"),
		}, {
			AttributeName: aws.String("Key"),
			KeyType:       aws.String("RANGE"),
		}},
		AttributeDefinitions: []*dynamodb.AttributeDefinition{{
			AttributeName: aws.String("Path"),
			AttributeType: aws.String("S"),
		}, {
			AttributeName: aws.String("Key"),
			AttributeType: aws.String("S"),
		}},
	}

	_, err = db.CreateTable(in)
	if err != nil {
		return err
	}

	Eventually(func() bool {
		resp, err := db.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: aws.String(table),
		})
		if err != nil {
			return false
		}

		return *resp.Table.TableStatus == "ACTIVE"

	}, 5*time.Minute, 3*time.Second).Should(BeTrue())

	return nil

}

func (f *Framework) DynamoDBDeleteTable(region, table string) error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	db := dynamodb.New(sess, aws.NewConfig().WithRegion(region))

	_, err = db.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(table),
	})

	return err
}
