const Http = new XMLHttpRequest();

function burnSubmit(token) {
    $("#submitBtn").attr("hidden", true);
    $("#submitText").fadeOut("slow",function() {
        $(this).text("...cooking binary!").fadeIn("slow");

        startBurn()
        Http.open("POST", 'http://localhost:8080/burn?token=' + token);
        Http.send();
        Http.onreadystatechange = (e) => {
            const res = Http.responseText;
            if (Http.status === 200 && Http.readyState === 4) {
                const resJson = JSON.parse(res);
                console.log(resJson)
                $("#submitText,#downloadRemote,#downloadLocal").fadeOut("slow",function() {
                    $(this).text("It's ready for download. Check below!").fadeIn("slow");
                    $("#downloadRemote").text("Download Remote").fadeIn("slow");
                    $("#downloadLocal").text("Download Local").fadeIn("slow");
                    $("#downloadLocal").attr("href", `http://localhost:8080/download/${resJson.fingerprint}/local`)
                    $("#downloadRemote").attr("href", `http://localhost:8080/download/${resJson.fingerprint}/remote`)
                    if (resJson.success){
                        stopBurn()
                    }
                })
            }
        }
    });

}
