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
                $("#submitText").fadeOut("slow",function() {
                    if (resJson.success){
                        stopBurn()
                        $(this).text("It's ready for download. Check above!").fadeIn("slow");
                        $("#allDownload").fadeIn("slow");
                        $("#downloadLocal").attr("href", `http://localhost:8080/download/${resJson.fingerprint}/floxyL`)
                        $("#downloadRemote").attr("href", `http://localhost:8080/download/${resJson.fingerprint}/floxyR`)
                    }else {
                        $(this).text("Sorry... some error occurred =(").fadeIn("slow");
                        stopBurn()
                    }
                })
            }
        }
    });

}
