const Http = new XMLHttpRequest();

$( document ).ready(function() {
    $("#home").hide();
    $("#burn").hide();
    $('#burning').hide();
    $('#sharePage').hide();
    if (window.location.pathname.includes('burn')) {
        $("#burn").show();
    }else if (window.location.pathname.includes('share')){
        const fingerprint = window.location.pathname.split('share/')[1].split('/')[0]
        $("#copyLocalLink").attr("href", `http://localhost:8080/api/download/${fingerprint}/floxyL`)
        $("#copyRemoteLink").attr("href", `http://localhost:8080/api/download/${fingerprint}/floxyR`)
        $("#sharePage").show();
        greenBurn();
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

const binaryExpOptions = ['1 hour', '1 day', '10 days', '1 month'];
let binaryExpSelected = -1;

function binaryExpSelector() {
    binaryExpSelected++;
    if (binaryExpSelected > binaryExpOptions.length -1){
        binaryExpSelected = 0;
    }
    $("#binaryExpSelector").text(binaryExpOptions[binaryExpSelected])
    enableBurnButton()
}

const localDistroOptions = ['Linux', 'Docker'];
let localDistroSelected = -1;

function localDistroSelector() {
    localDistroSelected++;
    if (localDistroSelected > localDistroOptions.length -1){
        localDistroSelected = 0;
    }
    $("#localDistroSelector").text(localDistroOptions[localDistroSelected])
    enableBurnButton()
}

const remoteDistroOptions = ['Linux', 'Docker'];
let remoteDistroSelected = -1;

function remoteDistroSelector() {
    remoteDistroSelected++;
    if (remoteDistroSelected > remoteDistroOptions.length -1){
        remoteDistroSelected = 0;
    }
    $("#remoteDistroSelector").text(remoteDistroOptions[remoteDistroSelected])
    enableBurnButton()
}

function enableBurnButton(){
    if (binaryExpSelected !== -1 && localDistroSelected !== -1 && remoteDistroSelected !== -1){
        $("#burnButtonContainer").fadeIn();
    }
}

function burnSubmit(token) {
    $("#burn").fadeOut("slow",function() {
        startBurn()
        $("#burning").fadeIn("slow",function() {
            Http.open("POST", 'http://localhost:8080/api/burn?token=' + token);
            Http.send();
            Http.onreadystatechange = (e) => {
                const res = Http.responseText;
                if (Http.status === 200 && Http.readyState === 4) {
                    const resJson = JSON.parse(res);
                    if (resJson.status === 'approved'){
                        window.history.pushState({}, 'Share floxy', `/share/${resJson.fingerprint}`);
                        $("#burning").fadeOut("slow",function() {
                            $("#copyLocalLink").attr("href", `http://localhost:8080/api/download/${resJson.fingerprint}/floxyL`)
                            $("#copyRemoteLink").attr("href", `http://localhost:8080/api/download/${resJson.fingerprint}/floxyR`)
                            $("#sharePage").fadeIn("slow");
                            greenBurn();
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
                    }
                }
            }
        })
    });
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