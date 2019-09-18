package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/roles"
)

type Customer struct {
	ID          uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `sql:"index"`
	Name        string
	Description string
}

func ConfigureQorResource(r resource.Resourcer) {
	// Configure resource with dummy Objects data structure

	var dummyCustomer1 Customer
	dummyCustomer1.ID, _ = uuid.Parse("1D50A411-4927-4812-B6D0-215E8620F68B")
	dummyCustomer1.Name = "dummy customer 1"
	dummyCustomer1.Description = "the first dummy customer"
	dummyCustomer1.CreatedAt = time.Now()
	dummyCustomer1.UpdatedAt = time.Now()

	var dummyCustomer2 Customer
	dummyCustomer2.ID, _ = uuid.Parse("0052B26D-CA72-434A-BAEF-8D047A2F9F32")
	dummyCustomer2.Name = "dummy customer 2"
	dummyCustomer2.Description = "the second dummy customer"
	dummyCustomer2.CreatedAt = time.Now()
	dummyCustomer2.UpdatedAt = time.Now()

	var dummyCustomer3 Customer
	dummyCustomer3.ID, _ = uuid.Parse("6400F6FA-56CA-457E-927B-CB18F44B298F")
	dummyCustomer3.Name = "dummy customer 3"
	dummyCustomer3.Description = "the third dummy customer"
	dummyCustomer3.CreatedAt = time.Now()
	dummyCustomer3.UpdatedAt = time.Now()

	dummyCustomers := make([]Customer, 0)
	dummyCustomers = append(dummyCustomers, dummyCustomer1)
	dummyCustomers = append(dummyCustomers, dummyCustomer2)
	dummyCustomers = append(dummyCustomers, dummyCustomer3)

	p, ok := r.(*admin.Resource)
	if !ok {
		panic(fmt.Sprintf("Unexpected resource! T: %T", r))
	}
	// find record and decode it to result
	p.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {

		if p.HasPermission(roles.Read, context) {

			var dummyCustomerTMP Customer
			fmt.Println("result before FindOneHandler: ", result)
			dummyCustomerTMP.ID, _ = uuid.Parse(context.ResourceID)
			for i := 0; i < len(dummyCustomers); i++ {
				if dummyCustomers[i].ID == dummyCustomerTMP.ID {
					var buf bytes.Buffer
					json.NewEncoder(&buf).Encode(dummyCustomers[i])
					json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&result)
				}
			}

			fmt.Println("result after FindOneHandler: ", result)

			return nil
		}

		return roles.ErrPermissionDenied
	}

	p.FindManyHandler = func(result interface{}, context *qor.Context) error {
		if p.HasPermission(roles.Read, context) {

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(dummyCustomers)
			json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&result)
			return nil
		}

		return roles.ErrPermissionDenied

	}

	p.SaveHandler = func(result interface{}, context *qor.Context) error {
		if p.HasPermission(roles.Create, context) || p.HasPermission(roles.Update, context) {
			tmpUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")

			var dummyCustomerTMP Customer

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(result)
			json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&dummyCustomerTMP)

			if dummyCustomerTMP.ID == tmpUUID {
				dummyCustomerTMP.ID, _ = uuid.NewRandom()
				dummyCustomers = append(dummyCustomers, dummyCustomerTMP)
			} else {
				for i := 0; i < len(dummyCustomers); i++ {
					if dummyCustomers[i].ID == dummyCustomerTMP.ID {
						var buf bytes.Buffer
						json.NewEncoder(&buf).Encode(dummyCustomerTMP)
						json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&dummyCustomers[i])
					}
				}
			}

			return nil
		}
		return roles.ErrPermissionDenied
	}

	p.DeleteHandler = func(result interface{}, context *qor.Context) error {
		if p.HasPermission(roles.Delete, context) {

			var dummyCustomerTMP Customer
			fmt.Println("result before DeleteHandler: ", result)
			dummyCustomerTMP.ID, _ = uuid.Parse(context.ResourceID)

			for i := 0; i < len(dummyCustomers); i++ {
				if dummyCustomers[i].ID == dummyCustomerTMP.ID {
					copy(dummyCustomers[i:], dummyCustomers[i+1:])
					dummyCustomers = dummyCustomers[:len(dummyCustomers)-1]
				}
			}

			return nil
		}
		return roles.ErrPermissionDenied
	}

}

func ConfigureQorResourceDynamoDB(r resource.Resourcer) {
	// Configure resource with DynamoDB

	config := &aws.Config{
		Region:   aws.String("us-west-2"),
		Endpoint: aws.String("http://localhost:8000"),
	}

	db := dynamo.New(session.New(), config)
	table := db.Table("Customers")

	p, ok := r.(*admin.Resource)
	if !ok {
		panic(fmt.Sprintf("Unexpected resource! T: %T", r))
	}

	p.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		fmt.Println("FindOneHandler")
		if p.HasPermission(roles.Read, context) {

			var dbcustomerTMP Customer
			fmt.Println("result before FindOneHandler: ", result)
			dbcustomerTMP.ID, _ = uuid.Parse(context.ResourceID)
			err := table.Get("ID", dbcustomerTMP.ID).One(&dbcustomerTMP)

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(dbcustomerTMP)
			json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&result)

			return err

		}

		return roles.ErrPermissionDenied
	}

	p.FindManyHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("FindManyHandler")
		if p.HasPermission(roles.Read, context) {

			var dbcustomers []Customer
			err := table.Scan().All(&dbcustomers)

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(dbcustomers)
			json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&result)

			return err
		}

		return roles.ErrPermissionDenied
	}

	p.SaveHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("SaveHandler")
		if p.HasPermission(roles.Create, context) || p.HasPermission(roles.Update, context) {

			tmpUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")

			var dummyCustomerTMP Customer

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(result)
			json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(&dummyCustomerTMP)

			var err error

			if dummyCustomerTMP.ID == tmpUUID {
				dummyCustomerTMP.ID, _ = uuid.NewRandom()
				err = table.Put(dummyCustomerTMP).Run()
			} else {
				err = table.Put(dummyCustomerTMP).Run()
			}

			return err

		}
		return roles.ErrPermissionDenied
	}

	p.DeleteHandler = func(result interface{}, context *qor.Context) error {
		fmt.Println("DeleteHandler")
		if p.HasPermission(roles.Delete, context) {
			var dbcustomerTMP Customer
			dbcustomerTMP.ID, _ = uuid.Parse(context.ResourceID)

			err := table.Delete("ID", dbcustomerTMP.ID).Run()

			return err
		}
		return roles.ErrPermissionDenied
	}

}
