// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package batch

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/batch/2017-09-01.6.0/batch"
	batchARM "github.com/Azure/azure-sdk-for-go/services/batch/mgmt/2017-09-01/batch"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	uuid "github.com/satori/go.uuid"
)

const (
	stdoutFile string = "stdout.txt"
	stderrFile string = "stderr.txt"
)

// CreateAzureBatchAccount creates a new azure batch account
func CreateAzureBatchAccount(ctx context.Context, accountName, location, resourceGroupName string) (a batchARM.Account, err error) {
	token, err := iam.GetResourceManagementToken(iam.AuthGrantType())

	if err != nil {
		return batchARM.Account{}, fmt.Errorf("cannot get auth token: %v", err)
	}

	accountClient := batchARM.NewAccountClient(helpers.SubscriptionID())
	accountClient.Authorizer = autorest.NewBearerAuthorizer(token)
	accountClient.AddToUserAgent(helpers.UserAgent())

	res, err := accountClient.Create(ctx, resourceGroupName, accountName, batchARM.AccountCreateParameters{
		Location: to.StringPtr(location),
	})

	if err != nil {
		return batchARM.Account{}, err
	}

	err = res.WaitForCompletion(ctx, accountClient.Client)

	if err != nil {
		return batchARM.Account{}, fmt.Errorf("failed waiting for account creation: %v", err)
	}

	account, err := res.Result(accountClient)

	if err != nil {
		return batchARM.Account{}, fmt.Errorf("failed retreiving for account: %v", err)
	}

	return account, nil
}

// CreateBatchPool creates an Azure Batch compute pool
func CreateBatchPool(ctx context.Context, accountName, accountLocation, poolID string) error {
	token, err := iam.GetBatchToken(iam.AuthGrantType())

	if err != nil {
		return fmt.Errorf("cannot get auth token: %v", err)
	}

	poolClient := batch.NewPoolClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	poolClient.Authorizer = autorest.NewBearerAuthorizer(token)
	poolClient.AddToUserAgent(helpers.UserAgent())
	poolClient.RequestInspector = fixContentTypeInspector()

	toCreate := batch.PoolAddParameter{
		ID: &poolID,
		VirtualMachineConfiguration: &batch.VirtualMachineConfiguration{
			ImageReference: &batch.ImageReference{
				Publisher: to.StringPtr("Canonical"),
				Sku:       to.StringPtr("16.04-LTS"),
				Offer:     to.StringPtr("UbuntuServer"),
				Version:   to.StringPtr("latest"),
			},
			NodeAgentSKUID: to.StringPtr("batch.node.ubuntu 16.04"),
		},
		MaxTasksPerNode:      to.Int32Ptr(1),
		TargetDedicatedNodes: to.Int32Ptr(1),
		// Create a startup task to run a script on each pool machine
		StartTask: &batch.StartTask{
			ResourceFiles: &[]batch.ResourceFile{
				batch.ResourceFile{
					BlobSource: to.StringPtr("https://gist.githubusercontent.com/lawrencegripper/795a9e809b52a0b8f874251c62a5a106/raw/8bfe4f4c80440204c994287e415c56f140dfd747/echohello.sh"),
					FilePath:   to.StringPtr("echohello.sh"),
					FileMode:   to.StringPtr("777"),
				},
			},
			CommandLine:    to.StringPtr("bash -f echohello.sh"),
			WaitForSuccess: to.BoolPtr(true),
			UserIdentity: &batch.UserIdentity{
				AutoUser: &batch.AutoUserSpecification{
					ElevationLevel: batch.Admin,
					Scope:          batch.Task,
				},
			},
		},
		VMSize: to.StringPtr("standard_a1"),
	}

	poolCreate, err := poolClient.Add(ctx, toCreate, nil, nil, nil, nil)

	if err != nil {
		return fmt.Errorf("cannot create pool: %v", err)
	}

	if poolCreate.StatusCode != 201 {
		log.Println(poolCreate)
		return errors.New("error creating pool, wasn't created")
	}

	return nil
}

// CreateBatchJob create an azure batch job
func CreateBatchJob(ctx context.Context, accountName, accountLocation, poolID, jobID string) error {
	token, err := iam.GetBatchToken(iam.AuthGrantType())

	if err != nil {
		return fmt.Errorf("cannot get auth token: %v", err)
	}

	jobClient := batch.NewJobClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	jobClient.Authorizer = autorest.NewBearerAuthorizer(token)
	jobClient.AddToUserAgent(helpers.UserAgent())
	jobClient.RequestInspector = fixContentTypeInspector()

	jobToCreate := batch.JobAddParameter{
		ID: to.StringPtr(jobID),
		PoolInfo: &batch.PoolInformation{
			PoolID: to.StringPtr(poolID),
		},
	}
	// reqID := uuid.NewV4()
	res, err := jobClient.Add(ctx, jobToCreate, nil, nil, nil, nil)

	if err != nil {
		return err
	}

	if res.StatusCode != 201 {
		log.Println(res)
		return errors.New("error creating job, wasn't created")
	}

	return nil
}

// CreateBatchTask create an azure batch job
func CreateBatchTask(ctx context.Context, accountName, accountLocation, jobID string) (string, error) {
	taskID := uuid.NewV4().String()
	taskClient := must(getTaskClient(accountName, accountLocation))
	taskToAdd := batch.TaskAddParameter{
		ID:          &taskID,
		CommandLine: to.StringPtr("/bin/bash -c 'set -e; set -o pipefail; echo Hello world from the Batch Hello world sample!; wait'"),
		UserIdentity: &batch.UserIdentity{
			AutoUser: &batch.AutoUserSpecification{
				ElevationLevel: batch.Admin,
				Scope:          batch.Task,
			},
		},
	}
	res, err := taskClient.Add(ctx, jobID, taskToAdd, nil, nil, nil, nil)

	if err != nil {
		return "", err
	}

	if res.StatusCode != 201 {
		log.Println(res)
		return "", errors.New("error creating job, wasn't created")
	}

	return taskID, nil
}

// WaitForTaskResult polls the task and retreives it's stdout once it has completed
func WaitForTaskResult(ctx context.Context, accountName, accountLocation, jobID, taskID string) (stdout string, err error) {
	taskClient := must(getTaskClient(accountName, accountLocation))
	res, err := taskClient.Get(ctx, jobID, taskID, "", "", nil, nil, nil, nil, "", "", nil, nil)
	if err != nil {
		return "", err
	}

	if res.State != batch.TaskStateCompleted {
		for {
			time.Sleep(time.Second * 15)
			res, err = taskClient.Get(ctx, jobID, taskID, "", "", nil, nil, nil, nil, "", "", nil, nil)
			if err != nil {
				return "", err
			}
			if res.State == batch.TaskStateCompleted {
				break
			}
		}
	}

	fileClient := batch.NewFileClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	fileClient.Authorizer = taskClient.Authorizer
	fileClient.AddToUserAgent(helpers.UserAgent())
	fileClient.RequestInspector = fixContentTypeInspector()

	reader, err := fileClient.GetFromTask(ctx, jobID, taskID, stdoutFile, nil, nil, nil, nil, "", nil, nil)

	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(*reader.Value)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func must(t *batch.TaskClient, err error) *batch.TaskClient {
	if err != nil {
		panic(err)
	}
	return t
}

func getTaskClient(accountName, accountLocation string) (*batch.TaskClient, error) {
	token, err := iam.GetBatchToken(iam.AuthGrantType())

	if err != nil {
		return &batch.TaskClient{}, fmt.Errorf("cannot get auth token: %v", err)
	}

	taskClient := batch.NewTaskClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	taskClient.Authorizer = autorest.NewBearerAuthorizer(token)
	taskClient.AddToUserAgent(helpers.UserAgent())
	taskClient.RequestInspector = fixContentTypeInspector()
	return &taskClient, nil
}

func getBatchBaseURL(accountName, accountLocation string) string {
	return fmt.Sprintf("https://%s.%s.batch.azure.com", accountName, accountLocation)
}

// This is required due to this issue: https://github.com/Azure/azure-sdk-for-go/issues/1159. Can be removed once resolved.
func fixContentTypeInspector() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			r.Header.Set("Content-Type", "application/json; odata=minimalmetadata")
			return r, nil
		})
	}
}
