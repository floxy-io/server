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

            $("#submitText").fadeOut("slow",function() {
                $(this).text("It's ready for download!").fadeIn("slow");
                if (res === "ok"){
                    stopBurn()
                }
            })
        }
    });

}
