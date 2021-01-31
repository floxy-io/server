const Http = new XMLHttpRequest();

$( document ).ready(function() {
    $("#home").hide();
    $("#burn").hide();
    $("#noLinkSharePage").hide();
    $('#burning').hide();
    $('#sharePage').hide();
    if (window.location.pathname.includes('burn')) {
        $("#burn").show();
    }else if (window.location.pathname.includes('share')){
        const fingerprint = window.location.pathname.split('share/')[1].split('/')[0]
        getFloxy(fingerprint)
    }else {
        $("#home").show();
    }
});

function toBurnpage() {
    window.history.pushState({}, 'Burn floxy', '/burn');
    $("#home").fadeOut("slow",function() {
        $("#burn").fadeIn("slow");
    })
}

const binaryExpOptions = [{text: '1 hour', value: 1}, {text: '1 day', value: 24}, {text: '10 days', value: 24 * 10}, {text: '1 month', value: 24 * 10 * 30}];

let binaryExpSelected = -1;

function binaryExpSelector() {
    binaryExpSelected++;
    if (binaryExpSelected > binaryExpOptions.length -1){
        binaryExpSelected = 0;
    }
    $("#binaryExpSelector").text(binaryExpOptions[binaryExpSelected].text)
    enableBurnButton()
}

const distroOptions = ['Linux'];
let distroSelected = -1;

function distroSelector() {
    distroSelected++;
    if (distroSelected > distroOptions.length -1){
        distroSelected = 0;
    }
    $("#distroSelector").text(distroOptions[distroSelected])
    enableBurnButton()
}

const remotePasswordOptions = ['No', 'Yes'];
let remotePasswordSelected = 0;

function remotePasswordSelector() {
    remotePasswordSelected++;
    if (remotePasswordSelected > remotePasswordOptions.length -1){
        remotePasswordSelected = 0;
    }
    $("#remotePasswordSelector").text(remotePasswordOptions[remotePasswordSelected])
    enableBurnButton()
}

function enableBurnButton(){
    if (binaryExpSelected !== -1 && distroSelected !== -1 && remotePasswordSelected !== -1){
        $("#burnButtonContainer").fadeIn();
    }
}

function burnSubmit(token) {
    $("#burn").fadeOut("slow",function() {
        startBurn()
        $("#burning").fadeIn("slow",function() {
            Http.open("POST", 'http://localhost:8080/api/floxy/burn');
            Http.setRequestHeader("Content-Type", "application/json");
            const data = JSON.stringify(
                {
                    "remotePassword": remotePasswordSelected === 1,
                    "expiration": binaryExpOptions[binaryExpSelected].value,
                    "token": token,
                }
            );
            Http.send(data);
            Http.onreadystatechange = (e) => {
                const res = Http.responseText;
                if (Http.status === 200 && Http.readyState === 4) {
                    const resJson = JSON.parse(res);
                    if (resJson.status === 'approved'){
                        window.history.pushState({}, 'Share floxy', `/share/${resJson.fingerprint}`);
                        $("#burning").fadeOut("slow",function() {
                            getFloxy(resJson.fingerprint)
                        })
                    }else if (resJson.status === 'non_approve'){
                        stopBurn()
                        $("#burnInProcess").fadeOut("slow",function() {
                            $("#burnNotApprove").fadeIn("slow")
                            stopBurn()
                        })
                    }else if (resJson.status === 'challenge'){
                        stopBurn()
                        $("#burnInProcess").fadeOut("slow",function() {
                            $("#burnChallenge").fadeIn("slow")
                            stopBurn()
                        })
                    }else {
                        stopBurn()
                        $("#burnInProcess").fadeOut("slow",function() {
                            $("#binGenerationError").fadeIn("slow")
                            stopBurn()
                        })
                    }
                }
            }
        })
    });
}

function getFloxy(fingerprint) {
    stopBurn();
    $("#copyLink").attr("href", `http://localhost:8080/api/download/${fingerprint}/floxy`)

    Http.open("GET", `http://localhost:8080/api/floxy/${fingerprint}`);

    Http.send();
    Http.onreadystatechange = (e) => {
        const res = Http.responseText;
        if (Http.readyState === 4) {
            if (Http.status === 200){
                const resJson = JSON.parse(res);
                if (resJson.remotePassword !== null && resJson.remotePassword !== ""){
                    $(".remoteCode").text(`remote>./floxy -k=remote -p=${resJson.remotePassword} -h {{host:ip}}`)
                }
                $("#shareExpSecurity").text(`${resJson.linkExpiration} minute(s)`)
                $("#sharePage").fadeIn("slow");
                // resJson.linkExpiration
            }else {
                $("#noLinkSharePage").fadeIn("slow");
            }

        }
    }
}

const getUrlParameter = function getUrlParameter(sParam) {
    let sPageURL = window.location.search.substring(1),
        sURLVariables = sPageURL.split('&'),
        sParameterName,
        i;

    for (i = 0; i < sURLVariables.length; i++) {
        sParameterName = sURLVariables[i].split('=');

        if (sParameterName[0] === sParam) {
            return sParameterName[1] === undefined ? true : decodeURIComponent(sParameterName[1]);
        }
    }
};
