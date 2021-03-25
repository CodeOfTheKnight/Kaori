const url = "https://127.0.0.1:8010/api/addData/music"
const url2 = "https://127.0.0.1:8010/api/test/files/addData/music/addMusic.json"
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.XbPfbIHMI6arZ3Y922BhjWgQzWXcXNrz0ogtVhfEd2o"

function getJson() {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4 && this.status == 200) {
            console.log(this.responseText)
            let data = this.responseText
            sendData(data)
        }
    };
    xhttp.open("GET", url2, true);
    xhttp.setRequestHeader("Authorization", "bearer "+token)
    //xhttp.setRequestHeader("Access-Control-Request-Method", "GET, POST, OPTION")
    //xhttp.setRequestHeader("Access-Control-Request-Headers", "*")
    //xhttp.setRequestHeader("Origin", "*")
    xhttp.send();
}

function sendData() {

    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4 && this.status == 200) {
            console.log(this.responseText)
        }
    };
    xhttp.open("POST", url, true)
   //xhttp.setRequestHeader("Authorization", "bearer " + token)
    xhttp.send(data);
}

const data = `{
  "idAnilist": 97832,
  "type": "OP",
  "numSong": 0,
  "isFull": true,
  "artist": "nano.RIPE",
  "nameSong": "Azalea",
  "cover": "data:image/jpeg;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAIAAAAmkwkpAAABhGlDQ1BJQ0MgcHJvZmlsZQAAKJF9kT1Iw0AcxV9TpSItgnYQcchQnSxIFXHUKhShQqgVWnUwufQLmhiSFBdHwbXg4Mdi1cHFWVcHV0EQ/ABxc3NSdJES/5cUWsR4cNyPd/ced+8AoVFlmtU1Dmi6bWZSSTGXXxFDrwiiHxGEkJCZZcxKUhq+4+seAb7exXmW/7k/R0QtWAwIiMQzzDBt4nXiqU3b4LxPHGVlWSU+Jx4z6YLEj1xXPH7jXHJZ4JlRM5uZI44Si6UOVjqYlU2NeJI4pmo65Qs5j1XOW5y1ao217slfGC7oy0tcpzmMFBawCAkiFNRQQRU24rTqpFjI0H7Sxz/k+iVyKeSqgJFjHhvQILt+8D/43a1VnEh4SeEk0P3iOB8jQGgXaNYd5/vYcZonQPAZuNLb/o0GMP1Jer2txY6Avm3g4rqtKXvA5Q4w+GTIpuxKQZpCsQi8n9E35YGBW6B31euttY/TByBLXaVvgINDYLRE2Ws+7+7p7O3fM63+fgBP23KZftE5qgAAAAlwSFlzAAAuIwAALiMBeKU/dgAAAAd0SU1FB+UDEg0JOMb1VdEAAAAZdEVYdENvbW1lbnQAQ3JlYXRlZCB3aXRoIEdJTVBXgQ4XAAAAMklEQVQI1wXBQQoAMQwCQLZUwXqKeM3/v7kz3+4maTszJ4ltkgBuW5LvPUl3ZgBIsv0DTPQChc5nh9gAAAAASUVORK5CYII=",
}`