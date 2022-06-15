package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	id        string = "id"
	item      string = "item"
	operation string = "operation"
	filename  string = "fileName"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func convertItemToUser(item []byte) (user User, err error) {
	if err = json.Unmarshal(item, &user); err != nil {
		return User{}, err
	}
	return user, nil
}

func unmarshalFile(f *os.File) (users []User, err error) {
	fileData, errRead := ioutil.ReadAll(f)
	if errRead != nil {
		return nil, errRead
	}

	if len(fileData) == 0 {
		return
	}

	if errUnmarshal := json.Unmarshal(fileData, &users); errUnmarshal != nil {
		return nil, errUnmarshal
	}

	return
}

func overWriteFile(f *os.File, users []User) (data []byte, err error) {
	data, errMarsh := json.Marshal(users)
	if errMarsh != nil {
		return nil, errMarsh
	}

	if errTrunc := f.Truncate(0); errTrunc != nil {
		return nil, errTrunc
	}
	if _, errSeek := f.Seek(0, 0); errSeek != nil {
		return nil, errSeek
	}

	_, err = f.WriteString(string(data))
	return
}

func (a Arguments) AddNewItem() ([]byte, error) {
	user, errConv := convertItemToUser([]byte(a[item]))
	if errConv != nil {
		return nil, errConv
	}

	f, err := os.OpenFile(a[filename], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	users, errCustomUnmarshal := unmarshalFile(f)
	if errCustomUnmarshal != nil {
		return nil, errCustomUnmarshal
	}

	for _, v := range users {
		if v.Id == user.Id {
			return []byte(fmt.Sprintf("Item with id %v already exists", user.Id)), nil
		}
	}

	users = append(users, user)

	return overWriteFile(f, users)
}

func (a Arguments) GettingListOfItem() ([]byte, error) {
	f, err := os.OpenFile(a[filename], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (a Arguments) RemoveUser() ([]byte, error) {
	f, err := os.OpenFile(a[filename], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	users, errCustomUnmarshal := unmarshalFile(f)
	if errCustomUnmarshal != nil {
		return nil, errCustomUnmarshal
	}

	var sameIDExist = false

	newUsers := []User{}

	for _, v := range users {
		if v.Id == a[id] {
			sameIDExist = true
		} else {
			newUsers = append(newUsers, v)
		}
	}

	if !sameIDExist {
		return []byte(fmt.Errorf("Item with id %v not found", a[id]).Error()), nil
	}

	return overWriteFile(f, newUsers)
}

func (a Arguments) FindById() ([]byte, error) {
	f, err := os.OpenFile(a[filename], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	users, errCustomUnmarshal := unmarshalFile(f)
	if errCustomUnmarshal != nil {
		return nil, errCustomUnmarshal
	}

	for _, v := range users {
		if v.Id == a[id] {
			return json.Marshal(v)
		}
	}

	return []byte(""), nil
}

func Perform(args Arguments, writer io.Writer) error {
	var ans []byte
	var err error

	if args[operation] == "" {
		return errors.New("-operation flag has to be specified")
	}

	if args[filename] == "" {
		return errors.New("-fileName flag has to be specified")
	}

	switch args[operation] {
	case "list":
		ans, err = args.GettingListOfItem()
		if err != nil {
			return err
		}

		writer.Write(ans)

		return nil
	case "add":
		if args[item] == "" {
			return errors.New("-item flag has to be specified")
		}

		res, err := args.AddNewItem()
		if err != nil {
			return err
		}

		writer.Write(res)

		return nil
	case "remove":
		if args[id] == "" {
			return errors.New("-id flag has to be specified")
		}

		ans, err = args.RemoveUser()
		if err != nil {
			return err
		}

		writer.Write(ans)

		return nil
	case "findById":
		if args[id] == "" {
			return errors.New("-id flag has to be specified")
		}

		ans, err = args.FindById()
		if err != nil {
			return err
		}

		writer.Write(ans)

		return nil
	default:
		return fmt.Errorf("Operation %s not allowed!", args[operation])
	}
}

func parseArgs() Arguments {
	id := flag.String("id", "", "id of user")
	item := flag.String("item", "", "data of user")
	operation := flag.String("operation", "", "what kind of operation do you need")
	fileName := flag.String("fileName", "", "which file do you want use")
	flag.Parse()

	return Arguments{"id": *id, "item": *item, "operation": *operation, "fileName": *fileName}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
