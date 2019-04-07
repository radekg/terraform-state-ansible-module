//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package account

import (
	"context"

	"strconv"

	"github.com/joyent/triton-go/account"
	tac "github.com/joyent/triton-go/account"
	"github.com/joyent/triton-go/cmd/config"
	"github.com/pkg/errors"
)

type AgentAccountClient struct {
	client *tac.AccountClient
}

func NewAccountClient(cfg *config.TritonClientConfig) (*AgentAccountClient, error) {
	accountClient, err := account.NewClient(cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "Error Creating Triton Account Client")
	}

	return &AgentAccountClient{
		client: accountClient,
	}, nil
}

func (c *AgentAccountClient) Get() (*tac.Account, error) {
	account, err := c.client.Get(context.Background(), &tac.GetInput{})
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (c *AgentAccountClient) UpdateAccount() (*tac.Account, error) {

	params := &tac.UpdateInput{}

	email := config.GetAccountEmail()
	if email != "" {
		params.Email = email
	}

	companyName := config.GetAccountCompanyName()
	if companyName != "" {
		params.CompanyName = companyName
	}

	firstName := config.GetAccountFirstName()
	if firstName != "" {
		params.FirstName = firstName
	}

	lastName := config.GetAccountLastName()
	if lastName != "" {
		params.LastName = lastName
	}

	address := config.GetAccountAddress()
	if address != "" {
		params.Address = address
	}

	postalCode := config.GetAccountPostalCode()
	if postalCode != "" {
		params.PostalCode = postalCode
	}

	city := config.GetAccountCity()
	if city != "" {
		params.City = city
	}

	state := config.GetAccountState()
	if state != "" {
		params.State = state
	}

	country := config.GetAccountCountry()
	if country != "" {
		params.Country = country
	}

	phone := config.GetAccountPhone()
	if phone != "" {
		params.Phone = phone
	}

	cnsEnabled := config.GetAccountCNSEnabled()
	if cnsEnabled != "" {
		b, _ := strconv.ParseBool(cnsEnabled)
		params.TritonCNSEnabled = b
	}

	account, err := c.client.Update(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (c *AgentAccountClient) ListKeys() ([]*tac.Key, error) {
	keys, err := c.client.Keys().List(context.Background(), &tac.ListKeysInput{})
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (c *AgentAccountClient) CreateKey() (*tac.Key, error) {
	params := &tac.CreateKeyInput{
		Key: config.GetSSHKey(),
	}

	name := config.GetSSHKeyName()
	if name != "" {
		params.Name = name
	}

	key, err := c.client.Keys().Create(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (c *AgentAccountClient) DeleteKey() (*tac.Key, error) {
	var key *tac.Key

	name := config.GetSSHKeyName()
	if name != "" {
		k, err := c.client.Keys().Get(context.Background(), &tac.GetKeyInput{
			KeyName: name,
		})
		if err != nil {
			return nil, err
		}

		key = k
	}

	fingerprint := config.GetSSHKeyFingerprint()
	if fingerprint != "" {
		keys, err := c.client.Keys().List(context.Background(), &tac.ListKeysInput{})
		if err != nil {
			return nil, err
		}

		for _, k := range keys {
			if k.Fingerprint == fingerprint {
				key = k
				break
			}
		}
	}

	err := c.client.Keys().Delete(context.Background(), &tac.DeleteKeyInput{
		KeyName: name,
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (c *AgentAccountClient) GetKey() (*tac.Key, error) {

	name := config.GetSSHKeyName()
	if name != "" {
		key, err := c.client.Keys().Get(context.Background(), &tac.GetKeyInput{
			KeyName: name,
		})
		if err != nil {
			return nil, err
		}

		return key, nil
	}

	fingerprint := config.GetSSHKeyFingerprint()
	if fingerprint != "" {
		keys, err := c.client.Keys().List(context.Background(), &tac.ListKeysInput{})
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			if key.Fingerprint == fingerprint {
				return key, nil
			}
		}
	}

	return nil, nil
}
