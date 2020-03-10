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
	"net/http"
	"time"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/batch/2017-05-01.5.0/batch"
	batchARM "github.com/Azure/azure-sdk-for-go/services/batch/mgmt/2017-09-01/batch"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	uuid "github.com/satori/go.uuid"
)

const (
	stdoutFile string = "stdout.txt"
)

func getAccountClient() batchARM.AccountClient {
	accountClient := batchARM.NewAccountClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	accountClient.Authorizer = auth
	accountClient.AddToUserAgent(config.UserAgent())
	return accountClient
}

func getPoolClient(accountName, accountLocation string) batch.PoolClient {
	poolClient := batch.NewPoolClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	auth, _ := iam.GetBatchAuthorizer()
	poolClient.Authorizer = auth
	poolClient.AddToUserAgent(config.UserAgent())
	poolClient.RequestInspector = fixContentTypeInspector()
	return poolClient
}

func getJobClient(accountName, accountLocation string) batch.JobClient {
	jobClient := batch.NewJobClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	auth, _ := iam.GetBatchAuthorizer()
	jobClient.Authorizer = auth
	jobClient.AddToUserAgent(config.UserAgent())
	jobClient.RequestInspector = fixContentTypeInspector()
	return jobClient
}

func getTaskClient(accountName, accountLocation string) batch.TaskClient {
	taskClient := batch.NewTaskClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	auth, _ := iam.GetBatchAuthorizer()
	taskClient.Authorizer = auth
	taskClient.AddToUserAgent(config.UserAgent())
	taskClient.RequestInspector = fixContentTypeInspector()
	return taskClient
}

func getFileClient(accountName, accountLocation string) batch.FileClient {
	fileClient := batch.NewFileClientWithBaseURI(getBatchBaseURL(accountName, accountLocation))
	auth, _ := iam.GetBatchAuthorizer()
	fileClient.Authorizer = auth
	fileClient.AddToUserAgent(config.UserAgent())
	fileClient.RequestInspector = fixContentTypeInspector()
	return fileClient
}

// CreateAzureBatchAccount creates a new azure batch account
func CreateAzureBatchAccount(ctx context.Context, accountName, location, resourceGroupName string) (a batchARM.Account, err error) {
	accountClient := getAccountClient()
	res, err := accountClient.Create(ctx, resourceGroupName, accountName, batchARM.AccountCreateParameters{
		Location: to.StringPtr(location),
	})

	if err != nil {
		return a, err
	}

	err = res.WaitForCompletionRef(ctx, accountClient.Client)

	if err != nil {
		return batchARM.Account{}, fmt.Errorf("failed waiting for account creation: %v", err)
	}

	account, err := res.Result(accountClient)

	if err != nil {
		return a, fmt.Errorf("failed retreiving for account: %v", err)
	}

	return account, nil
}

// CreateBatchPool creates an Azure Batch compute pool
func CreateBatchPool(ctx context.Context, accountName, accountLocation, poolID string) error {
	poolClient := getPoolClient(accountName, accountLocation)
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
				{
					BlobSource: to.StringPtr("https://raw.githubusercontent.com/lawrencegripper/azure-sdk-for-go-samples/1441a1dc4a6f7e47c4f6d8b537cf77ce4f7c452c/batch/examplestartup.sh"),
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

	_, err := poolClient.Add(ctx, toCreate, nil, nil, nil, nil)

	if err != nil {
		return fmt.Errorf("cannot create pool: %v", err)
	}

	return nil
}

// CreateBatchJob create an azure batch job
func CreateBatchJob(ctx context.Context, accountName, accountLocation, poolID, jobID string) error {
	jobClient := getJobClient(accountName, accountLocation)
	jobToCreate := batch.JobAddParameter{
		ID: to.StringPtr(jobID),
		PoolInfo: &batch.PoolInformation{
			PoolID: to.StringPtr(poolID),
		},
	}
	_, err := jobClient.Add(ctx, jobToCreate, nil, nil, nil, nil)

	if err != nil {
		return err
	}

	return nil
}

// CreateBatchTask create an azure batch job
func CreateBatchTask(ctx context.Context, accountName, accountLocation, jobID string) (string, error) {
	taskID := uuid.NewV4().String()
	taskClient := getTaskClient(accountName, accountLocation)
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
	_, err := taskClient.Add(ctx, jobID, taskToAdd, nil, nil, nil, nil)

	if err != nil {
		return "", err
	}

	return taskID, nil
}

// WaitForTaskResult polls the task and retreives it's stdout once it has completed
func WaitForTaskResult(ctx context.Context, accountName, accountLocation, jobID, taskID string) (stdout string, err error) {
	taskClient := getTaskClient(accountName, accountLocation)
	res, err := taskClient.Get(ctx, jobID, taskID, "", "", nil, nil, nil, nil, "", "", nil, nil)
	if err != nil {
		return "", err
	}
	waitCtx, cancel := context.WithTimeout(ctx, time.Minute*4)
	defer cancel()

	if res.State != batch.TaskStateCompleted {
		for {
			_, ok := waitCtx.Deadline()
			if !ok {
				return stdout, errors.New("timedout waiting for task to execute")
			}
			time.Sleep(time.Second * 15)
			res, err = taskClient.Get(ctx, jobID, taskID, "", "", nil, nil, nil, nil, "", "", nil, nil)
			if err != nil {
				return "", err
			}
			if res.State == batch.TaskStateCompleted {
				waitCtx.Done()
				break
			}
		}
	}

	fileClient := getFileClient(accountName, accountLocation)

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

func getBatchBaseURL(accountName, accountLocation string) string {
	return fmt.Sprintf("https://%s.%s.batch.azure.com", accountName, accountLocation)
}

// This is required due to this issue: https://github.com/Azure/azure-sdk-for-go/issues/1159. Can be removed once resolved.
func fixContentTypeInspector() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			r, err := p.Prepare(r)
			if err == nil {
				r.Header.Set("Content-Type", "application/json; odata=minimalmetadata")
			}
			return r, nil
		})
	}
}
