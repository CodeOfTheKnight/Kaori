package main

type endpoints string

const (
	endpointRoot  endpoints = "/"
	endpointGui   endpoints = "/KaoriGui/"
	endpointLogin endpoints = "/login"

	//Endpoint per servizio
	endpointService endpoints = "/api/service/"
	serviceAnime    endpoints = "/anime"   //[GET] Parametri: IdAnilist
	serviceManga    endpoints = "/manga"   //[GET] Parametri: IdAnilist
	serviceChapter  endpoints = "/chapter" //[GET] Parametri: IdAnilist, numEpisodio, sito
	serviceEpisode  endpoints = "/episode" //[GET] Parametri: IdAnilist, numCapitolo, sito
	serviceMusic    endpoints = "/music"   //[GET] Parametri: IdAnilist, [tipo](OP, ED, OST)

	//Endpoint utenti
	endpointUser endpoints = "/api/user/"
	userInfo     endpoints = "/info"     //[GET]
	userSettings endpoints = "/settings" //[GET]
	userBadge    endpoints = "/badge"    //[GET]

	//Endpoint auth
	endpointAuth      endpoints = "/api/auth/"
	authRefresh       endpoints = "/refresh"
	authLogin         endpoints = "/login"
	authSignUp        endpoints = "/signup"
	authConfirmSignUp endpoints = "/confirm"
	authUserExist     endpoints = "/exist" //[GET] Parametri: email

	//Endpoint addData
	endpointAddData endpoints = "/api/addData/"
	addDataAnime    endpoints = "/anime" //[POST] Struttura dati da definire.
	addDataManga    endpoints = "/manga" //[POST] Struttura dati da definire.
	addDataMusic    endpoints = "/music" //[POST] Struttura dati da definire.

	//Endpoint lista anime/manga
	endpointUserList endpoints = "/api/list/userList" //[Get] Parametri: TokenAnilist, TokenServer, TipoLista
	endpointUserInfo endpoints = "/api/list/userinfo" //[Get] Parametri: TokenAnilist, TokenServer

	//Endpoint for tests
	endpointTest endpoints = "/api/test/"
	testFiles    endpoints = "/files/"
)

func (e endpoints) String() string {
	return string(e)
}
