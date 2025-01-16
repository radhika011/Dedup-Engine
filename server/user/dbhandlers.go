package user

import (
	"dedup-server/mongoutil"
	"dedup-server/types"
)

func RegisterDBHandler(user types.UserData) (Response, error) {
	// funcName := "RegisterDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("UserData")

	response := Response{
		Status:   -1,
		UserData: types.UserData{},
	}

	userExists, _, err := mongoutil.FindUserData(collection, user.EmailID)

	if err != nil {
		return response, err
	}

	if userExists {
		response.Status = 3
		return response, nil
	}

	err = mongoutil.InsertUserData(collection, user)

	if err != nil {
		return response, err
	}

	response.Status = 0
	return response, nil
}

func VerifyDBHandler(user types.UserData) (Response, error) {
	// funcName := "VerifyDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("UserData")

	response := Response{
		Status:   -1,
		UserData: types.UserData{},
	}

	userExists, result, err := mongoutil.FindUserData(collection, user.EmailID)

	if err != nil {
		return response, err
	}

	if !userExists {
		response.Status = 2
		return response, nil
	}

	if result.Password != user.Password {
		response.Status = 1
		return response, nil
	}

	response.Status = 0
	return response, nil
}

func LoginDBHandler(user types.UserData) (Response, error) {
	// funcName := "LoginDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("UserData")

	response := Response{
		Status:   -1,
		UserData: types.UserData{},
	}

	userExists, result, err := mongoutil.FindUserData(collection, user.EmailID)

	if err != nil {
		return response, err
	}

	if !userExists {
		response.Status = 2
		return response, nil
	}

	if result.Password != user.Password {
		response.Status = 1
		return response, nil
	}

	response.Status = 0
	result.Password = [32]byte{}
	response.UserData = result
	return response, nil
}

func UpdateDBHandler(user types.UserData) (Response, error) {
	// funcName := "UpdateDBHandler"
	collection := mongoutil.Client.Database(mongoutil.DB_NAME).Collection("UserData")

	response := Response{
		Status:   -1,
		UserData: types.UserData{},
	}

	updated, err := mongoutil.UpdateUserData(collection, user)

	if err != nil {
		return response, err
	}

	if !updated {
		response.Status = 2
		return response, nil
	}

	response.Status = 0
	return response, nil
}
