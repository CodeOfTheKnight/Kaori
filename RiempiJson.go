package main

import (
    "github.com/KiritoNya/animeworld_V2/animeworld"
    "fmt"
    "strings"
    "io/ioutil"
    "encoding/json"
)

type Anime struct {
    Nome string `json:"nome"`
    LinkAnimeworld string `json:"link_animeworld"`
    Copertina string `json:"copertina"`
    IdAnilist string `json:id_anilist`
    
}


func main() {
    
    var a []Anime
    var idAnilist string
    
    series := animeworld.GetAllSeries()
    for _, serie := range series {
        nome := animeworld.GetNomeSerie(animeworld.BaseUrl + serie)
        copertina := animeworld.GetCopertina(animeworld.BaseUrl + serie)
        linkAnilist := animeworld.GetLinkAnilist(animeworld.BaseUrl + serie)
        if linkAnilist != "" {
            idAnilistMatrix := strings.Split(linkAnilist, "/")
            idAnilist = idAnilistMatrix[4]
        }
        a = append(a, Anime{Nome: nome, LinkAnimeworld: serie, Copertina: copertina, IdAnilist: idAnilist})
        fmt.Println("Aggiunta di " + nome + " nel database...")
        
        //Creazione JSON
        b, err := json.MarshalIndent(a, "", "\t")
        if err != nil {
            fmt.Println("error:", err)
        }
        err = ioutil.WriteFile("test.json", b, 0644)
        if err != nil {
            fmt.Println(err)
        }
    }
}
