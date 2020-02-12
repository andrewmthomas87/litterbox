package graphql

import (
	"encoding/json"
	"github.com/andrewmthomas87/litterbox/graphql/models"
	"github.com/vektah/gqlparser/gqlerror"
	"regexp"
)

var (
	validName    = regexp.MustCompile(`^[a-zA-Z]+(([',. -][a-zA-Z ])?[a-zA-Z]*)*$`)
	validAddress = regexp.MustCompile(`^[#.0-9a-zA-Z\s,-]+$`)
)

func validateInformation(information models.InformationInput) error {
	var errors models.InformationErrors
	isError := false

	if !validName.Match([]byte(information.Name)) {
		errors.Name = "Invalid name"
		isError = true
	}

	if information.OnCampus {
		if _, ok := models.BuildingLookup[information.Building]; !ok {
			errors.Building = "Invalid building"
			isError = true
		}
	} else {
		if !validAddress.Match([]byte(information.Address)) {
			errors.Address = "Invalid address"
			isError = true
		}
	}

	if isError {
		message, err := json.Marshal(errors)
		if err != nil {
			return err
		}

		return &gqlerror.Error{Message: string(message)}
	}

	return nil
}
