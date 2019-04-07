//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"time"

	"github.com/joyent/triton-go/client"
	"github.com/joyent/triton-go/errors"
	pkgerrors "github.com/pkg/errors"
)

type UsersClient struct {
	Client *client.Client
}

type User struct {
	ID           string    `json:"id"`
	Login        string    `json:"login"`
	EmailAddress string    `json:"email"`
	CompanyName  string    `json:"companyName"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Address      string    `json:"address"`
	PostalCode   string    `json:"postCode"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	Phone        string    `json:"phone"`
	Roles        []string  `json:"roles"`
	DefaultRoles []string  `json:"defaultRoles"`
	CreatedAt    time.Time `json:"created"`
	UpdatedAt    time.Time `json:"updated"`
}

type ListUsersInput struct{}

func (c *UsersClient) List(ctx context.Context, _ *ListUsersInput) ([]*User, error) {
	fullPath := path.Join("/", c.Client.AccountName, "users")
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to list users")
	}

	var result []*User
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode list users response")
	}

	return result, nil
}

type GetUserInput struct {
	UserID string
}

func (c *UsersClient) Get(ctx context.Context, input *GetUserInput) (*User, error) {
	fullPath := path.Join("/", c.Client.AccountName, "users", input.UserID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to get user")
	}

	var result *User
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode get user response")
	}

	return result, nil
}

type DeleteUserInput struct {
	UserID string
}

func (c *UsersClient) Delete(ctx context.Context, input *DeleteUserInput) error {
	fullPath := path.Join("/", c.Client.AccountName, "users", input.UserID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   fullPath,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return pkgerrors.Wrap(err, "unable to delete user")
	}

	return nil
}

type CreateUserInput struct {
	Email       string `json:"email"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	CompanyName string `json:"companyName,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Address     string `json:"address,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

func (c *UsersClient) Create(ctx context.Context, input *CreateUserInput) (*User, error) {
	fullPath := path.Join("/", c.Client.AccountName, "users")
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to create user")
	}

	var result *User
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode create user response")
	}

	return result, nil
}

type UpdateUserInput struct {
	UserID      string
	Email       string `json:"email,omitempty"`
	Login       string `json:"login,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Address     string `json:"address,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

func (c *UsersClient) Update(ctx context.Context, input *UpdateUserInput) (*User, error) {
	fullPath := path.Join("/", c.Client.AccountName, "users", input.UserID)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to update user")
	}

	var result *User
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode update user response")
	}

	return result, nil
}

// ChangeUserPasswordInput represents the options that can be specified
// when changing an existing user password.
type ChangeUserPasswordInput struct {
	// The unique ID of the user for which to change the password. Required.
	UserID string
	// User password. Required.
	Password string `json:"password"`
	// User password confirmation. Required.
	PasswordConfirmation string `json:"password_confirmation"`
}

// ChangeUserPassword updates password for an existing user. Returns updated
// user object, or an error otherwise.
func (c *UsersClient) ChangeUserPassword(ctx context.Context, input *ChangeUserPasswordInput) (*User, error) {
	fullPath := path.Join("/", c.Client.AccountName, "users", input.UserID, "change_password")

	// Verify both values locally, and save on an unnecessary API call,
	// but should the values be different, then return the same error
	// as what the CloudAPI would.
	if input.Password != input.PasswordConfirmation {
		return nil, pkgerrors.New("password and password confirmation must have the same value")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		// The `passwordInHistory` error message originates directly
		// from Triton's UFDS service through the CloudAPI, but it's
		// not very informative to the user.
		apiError, ok := pkgerrors.Cause(err).(*errors.APIError)
		if ok && apiError.Message == "passwordInHistory" {
			err = &errors.APIError{
				StatusCode: apiError.StatusCode,
				Code:       apiError.Code,
				Message:    "previous password cannot be reused",
			}
		}
		return nil, pkgerrors.Wrap(err, "unable to change user password")
	}

	var result *User
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode change user password response")
	}

	return result, nil
}
