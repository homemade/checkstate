package checkstate

import (
	"embed"
	"path"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mjarkk/mongomock"
)

//go:embed all:testdata
var TestData embed.FS

const TestDataPath = "testdata"

type MockContact struct {
	ModelWithState
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	FirstName string             `bson:"firstName"`
	LastName  string             `bson:"lastName"`
	Email     string             `bson:"email"`
}

func TestStateMachine(t *testing.T) {

	initialContact := MockContact{
		FirstName: "Tim",
		LastName:  "Test",
		Email:     "tim.test@example.org",
	}

	// create permitted state transitions using YAML based DSL
	yamlFile := path.Join(TestDataPath, "example.checkstate.yaml")
	data, err := TestData.ReadFile(yamlFile)
	if err != nil {
		t.Fatal(err)
	}
	var permittedStateTransitions PermittedStateTransitions
	permittedStateTransitions, err = PermittedStateTransitionsFromYAML(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("succesfully parsed the following permitted state transitions from %s %#v", yamlFile, permittedStateTransitions)

	// try and create a state machine in an initial invalid state - should return an error
	initialContact.State = "INVALID_STATE"
	_, err = CreateStateMachine(&initialContact, permittedStateTransitions)
	if err == nil {
		t.Fatal("created a state machine in an initial invalid state")
	}

	// reset state to create a state machine in a valid initial state
	initialContact.State = "CREATED"
	var stateMachine StateMachine
	stateMachine, err = CreateStateMachine(&initialContact, permittedStateTransitions)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("succesfully created state machine from initial contact %#v", stateMachine)

	// create mongo document containing embedded state machine
	db := mongomock.NewDB()
	collection := db.Collection("contacts")
	err = collection.Insert(initialContact)
	if err != nil {
		t.Fatal(err)
	}

	// retrieve created mongo document
	retrievedContact := MockContact{}
	err = collection.FindFirst(&retrievedContact, primitive.M{"email": "tim.test@example.org"})
	if err != nil {
		t.Fatal(err)
	}
	if retrievedContact.Email != initialContact.Email {
		t.Fatal("failed to retrieve created mongo document")
	}
	if retrievedContact.State != "CREATED" {
		t.Fatal("retrieved contact's initial state does not match expected value")
	}

	stateMachine, err = CreateStateMachine(&retrievedContact, permittedStateTransitions)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("succesfully created state machine from retrieved contact %#v", stateMachine)

	// update contact
	retrievedContact.FirstName = "Tony"
	err = stateMachine.Move(Event{
		Timestamp:   time.Now(),
		Description: "Changed FirstName",
	}, "UPDATED")
	if err != nil {
		t.Fatal(err)
	}
	err = collection.ReplaceFirstByID(retrievedContact.Id, retrievedContact)
	if err != nil {
		t.Fatal(err)
	}

	// retrieve updated mongo document
	updatedContact := MockContact{}
	err = collection.FindFirst(&updatedContact, primitive.M{"email": "tim.test@example.org"})
	if err != nil {
		t.Fatal(err)
	}
	if retrievedContact.Email != initialContact.Email {
		t.Fatal("failed to retrieve updated mongo document")
	}
	if retrievedContact.State != "UPDATED" {
		t.Fatal("updated contact's state does not match expected value")
	}

	// move to final state
	stateMachine, err = CreateStateMachine(&retrievedContact, permittedStateTransitions)
	if err != nil {
		t.Fatal(err)
	}
	err = stateMachine.Move(Event{
		Timestamp:   time.Now(),
		Description: "Finished editing",
	}, "COMPLETED")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("final state of contact %#v", retrievedContact)

}
