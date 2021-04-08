package main

import "errors"

type Permission string

const(
	UserPerm Permission = "u"
	CreatorPerm Permission = "c"
	TesterPerm Permission = "t"
	AdminPerm Permission = "a"
)

func (p Permission) ToString() string {
	return string(p)
}

func (jwt *JWTAccessMetadata) GetPermission() (perms []Permission, err error) {

	runes := []rune(jwt.Permission)

	for _, r := range runes {
		p, err := runeToPermission(r)
		if err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}

	return perms, nil
}

func runeToPermission(r rune) (Permission, error) {
	switch(string(r)){
	case UserPerm.ToString(): return UserPerm, nil
	case CreatorPerm.ToString(): return CreatorPerm, nil
	case TesterPerm.ToString(): return TesterPerm, nil
	case AdminPerm.ToString(): return AdminPerm, nil
	default: return Permission(""), errors.New("Error, permission not setted!")
	}
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
		if !myFunc(perms, preq){
			return false
		}
	}

	return true
}