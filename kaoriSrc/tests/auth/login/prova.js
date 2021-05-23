function prova() {
     var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
       console.log(this.response.headers);
    }
    };
    xhttp.open("POST", "https://waifuai.dd-dns.de:8012/api/auth/login", true);
    xhttp.send(`{"email": "diego.brignoli@gmail.com","password": "prova"}`);
}
