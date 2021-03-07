searchAnime("Citrus");
getEpisodes("Citrus");

function searchAnime(name) {
        
        var xhttp = new XMLHttpRequest();

        var myObj = { nome: name, lingua: "romaji", pageSize: 1000000 };
        var myJSON = JSON.stringify(myObj);
        console.log(myJSON);

        xhttp.open("POST", "http://127.0.0.1:8000/search", true);
        xhttp.send(myJSON);

        xhttp.onreadystatechange = function() {
            if (this.readyState == 4 && this.status == 200) {
                var response = this.responseText;
                console.log(response);
                //var respJson = JSON.parse(response); ESEMPIO
                //respJson.forEach(creaLocandina); ESEMPIO
            }
        };
}


/*ESEMPIO

function creaLocandina(item, index) {
            
    var campo = document.getElementById("campo");
    var div = document.createElement("div");
    var p = document.createElement("p");
    var textNode = document.createTextNode(item.nome);
    var img = document.createElement("img");
    img.src = item.urlImage.medium;
            
    p.appendChild(textNode);
    div.appendChild(img);
    div.appendChild(p);
    campo.appendChild(div);

}
*/

function getEpisodes(nome) {        
        
        var xhttp = new XMLHttpRequest();

        xhttp.open("POST", "http://127.0.0.1:8000/episode", true);
        xhttp.send(nome); //nome è una string semplice come "Citrus"

        xhttp.onreadystatechange = function() {
            if (this.readyState == 4 && this.status == 200) {
                var response = this.responseText;
                console.log(response)
                return response;
            }
        };

}
