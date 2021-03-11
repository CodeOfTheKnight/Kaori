package main

type endpoints string

const (
	endpointRoot endpoints = "/"
	endpointGui  endpoints = "/KaoriGui/"

	//Endpoint per servizio
	endApiInfo endpoints = "/api/info" //[GET] Parametri: IdAnilist
	endApiAnime endpoints = "/api/anime" //[GET] Parametri: IdAnilist
	endApiManga endpoints = "/api/manga" //[GET] Parametri: IdAnilist
	endApiChapter endpoints = "/api/chapter" //[GET] Parametri: IdAnilist, numEpisodio, sito
	endApiEpisode endpoints = "/api/episode" //[GET] Parametri: IdAnilist, numCapitolo, sito
	endApiMuisc endpoints = "/api/music" //[GET] Parametri: IdAnilist, [tipo](OP, ED, OST)

	//Endpoint di gestione
	endApiSignUp endpoints = "/api/signup" //[POST] Parametri: Username, mail, password.
	endApiLogin endpoints = "/api/login" //[POST] Parametri: mail, password  | Utilizza basic auth + JWT

	//Endpoint addData
	endApiAddDataAnime endpoints = "/api/addData/anime" //[POST] Struttura dati da definire.
	endApiAddDataManga endpoints = "/api/addData/manga" //[POST] Struttura dati da definire.
	endApiAddDataMusic endpoints = "/api/addData/music" //[POST] Struttura dati da definire.

	//Endpoint lista anime/manga
	endpointUserList endpoints = "/api/list/userList" //[Get] Parametri: TokenAnilist, TokenServer, TipoLista
	endpointUserInfo endpoints = "/api/list/userinfo" //[Get] Parametri: TokenAnilist, TokenServer

)

func (e endpoints) String() string {
	return string(e)
}