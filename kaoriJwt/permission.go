package kaoriJwt

import (
	"errors"
	"github.com/CodeOfTheKnight/Kaori/kaoriDatabase"
)

type Permission string

/*PRIVILEGI: ucta
u: User = Accesso base
c: Creator = Accesso alle api di verifica dati aggiunti da utenti
t: Tester = Accesso alle api di test
a: Admin = tutti gli accessi
*/

const (
	UserPerm    Permission = "u"
	CreatorPerm Permission = "c"
	TesterPerm  Permission = "t"
	AdminPerm   Permission = "a"
)

func runeToPermission(r rune) (Permission, error) {
	switch string(r) {
	case UserPerm.ToString():
		return UserPerm, nil
	case CreatorPerm.ToString():
		return CreatorPerm, nil
	case TesterPerm.ToString():
		return TesterPerm, nil
	case AdminPerm.ToString():
		return AdminPerm, nil
	default:
		return Permission(""), errors.New("Error, permission not setted!")
	}
}

func (p Permission) ToString() string {
	return string(p)
}

func GetPermissionFromDB(db *kaoriDatabase.NoSqlDb, email string) (Permission, error) {

	//Get permissions
	document, err := db.Client.C.Collection("User").Doc(email).Get(db.Client.Ctx)
	if err != nil {
		return "", err
	}
	data := document.Data()

	permString := data["Permission"].(string)
	return Permission(permString), nil
}

func IsAuthorized(perms []Permission, permsRequire ...Permission) bool {

	myFunc := func(permss []Permission, permReq Permission) bool {
		for _, p := range permss {
			if p == permReq {
				return true
			}
		}
		return false
	}

	for _, preq := range permsRequire {
		if !myFunc(perms, preq) {
			return false
		}
	}

	return true
}
