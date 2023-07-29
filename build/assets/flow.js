const Http = new XMLHttpRequest();

$( document ).ready(function() {
    flow();
});

window.onpopstate = function(event) {
    flow();
};

function flow() {
    $("#home").hide();
    $("#aboutPage").hide();
    $("#formPage").hide();
    $("#burnPage").hide();
    $("#noLinkSharePage").hide();
    $('#burningPage').hide();
    $('#sharePage').hide();
    if (window.location.pathname.includes('burn')) {
        $("#burnPage").show();
    }else if (window.location.pathname.includes('share')) {
        const fingerprint = window.location.pathname.split('share/')[1].split('/')[0]
        getFloxy(fingerprint)
    }else if (window.location.pathname.includes('about')) {
        $("#canvas").fadeTo( "slow" , 0.5)
        $("#aboutPage").show();
    }else if (window.location.pathname.includes('form')) {
        $("#formPage").show();
    }else {
        $("#home").show();
    }
}

function goToFormPage() {
    $("#canvas").fadeTo( "slow" , 0.5)
    window.history.pushState({}, 'Form', '/form');
    $("#burnPage").fadeOut("slow",function() {
        $("#formPage").fadeIn("slow");
    })
}

function toBurnpage() {
    $("#canvas").fadeTo( "slow" , 1)
    window.history.pushState({}, 'Burn floxy', '/burn');
    $("#home").fadeOut("slow",function() {
        $("#burnPage").fadeIn("slow");
    })
}

function toAboutpage() {
    $("#canvas").fadeTo( "slow" , 0.5)
    window.history.pushState({}, 'About', '/about');
    $("#home").fadeOut("slow",function() {
        $("#aboutPage").fadeIn("slow");
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
    if (canSubmit()){
        $("#burnButton").removeClass("disabled")
    }
}

const distroOptions = [{os: 'linux', platform: 'amd64', distro: 'Linux amd64'},{os: 'darwin',platform: 'amd64', distro: 'Darwin amd64'},{os: 'windows', platform: 'amd64', distro: 'Win amd64'}];
let localDistroSelected = -1;

function localDistroSelector() {
    localDistroSelected++;
    if (localDistroSelected > distroOptions.length -1){
        localDistroSelected = 0;
    }
    $("#localDistroSelector").text(distroOptions[localDistroSelected].distro)
    if (canSubmit()){
        $("#burnButton").removeClass("disabled")
    }
}

let remoteDistroSelected = -1;

function remoteDistroSelector() {
    remoteDistroSelected++;
    if (remoteDistroSelected > distroOptions.length -1){
        remoteDistroSelected = 0;
    }
    $("#remoteDistroSelector").text(distroOptions[remoteDistroSelected].distro)
    if (canSubmit()){
        $("#burnButton").removeClass("disabled")
    }
}

const remotePasswordOptions = ['No', 'Yes'];
let remotePasswordSelected = 0;

function remotePasswordSelector() {
    remotePasswordSelected++;
    if (remotePasswordSelected > remotePasswordOptions.length -1){
        remotePasswordSelected = 0;
    }
    $("#remotePasswordSelector").text(remotePasswordOptions[remotePasswordSelected])
    if (canSubmit()){
        $("#burnButton").removeClass("disabled")
    }
}

function canSubmit(){
    if (binaryExpSelected !== -1 && localDistroSelected !== -1 && remoteDistroSelected !== -1){
        return true;
    }
    return false;
}

function submit() {
    grecaptcha.ready(function() {
        grecaptcha.execute('6LeUMCMaAAAAABmP3FZtGTOFcGDgpGR0Z0pI7j2R', {action: 'submit'}).then(function(token) {
            burnSubmit(token)
        });
    });
}

function burnSubmit(token) {
    if (!canSubmit()){
        return
    }
    $("#canvas").fadeTo( "slow" , 1)
    $("#formPage").fadeOut("slow",function() {
        startBurn()
        $("#burningPage").fadeIn("slow",function() {
            Http.open("POST", 'https://floxy.io/api/floxy/burn');
            Http.setRequestHeader("Content-Type", "application/json");
            const data = JSON.stringify(
                {
                    "remotePassword": remotePasswordSelected === 1,
                    "expiration": binaryExpOptions[binaryExpSelected].value,
                    "distro": [
                        {'kind': 'local', 'os': distroOptions[localDistroSelected].os, 'platform': distroOptions[localDistroSelected].platform},
                        {'kind': 'remote', 'os': distroOptions[remoteDistroSelected].os, 'platform': distroOptions[remoteDistroSelected].platform}
                    ],
                    "token": token,
                }
            );
            Http.send(data);
            Http.onreadystatechange = (e) => {
                const res = Http.responseText;
                if (Http.status === 200 && Http.readyState === 4) {
                    const resJson = JSON.parse(res);
                    if (resJson.status === 'burning'){
                        window.history.pushState({}, 'Share floxy', `/share/${resJson.fingerprint}`);
                        getFloxy(resJson.fingerprint)
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
    // $(".copyLink").attr("href", `https://floxy.io/api/download/${fingerprint}/floxy`)

    Http.open("GET", `https://floxy.io/api/floxy/${fingerprint}`);
    Http.send();
    Http.onreadystatechange = (e) => {
        const res = Http.responseText;
        if (Http.readyState === 4) {
            if (!getFloxyResult(Http.status, res)){
                iterateUntilResult(fingerprint)
            }
        }
    }
}

function iterateUntilResult(fingerprint){
    const loop = setInterval(function(){
        Http.open("GET", `https://floxy.io/api/floxy/${fingerprint}`);
        Http.send();
        Http.onreadystatechange = (e) => {
            const res = Http.responseText;
            if (Http.readyState === 4) {
                if (getFloxyResult(Http.status, res)){
                    clearInterval(loop)
                }
            }
        }
    }, 3000);
}

function getFloxyResult(status, res) {
    if (status === 200) {
        const resJson = JSON.parse(res);
        if (resJson.status === 'active') {
            $("#burningPage").fadeOut("slow",function() {
                $("#canvas").fadeTo( "slow" , 0.5)
                stopBurn();
                if (resJson.remotePassword !== null && resJson.remotePassword !== "") {
                    $(".remoteCode").text(`remote>./floxyL -k=remote -p=${resJson.remotePassword} -h {{host:ip}}`)
                }
                $("#shareExpSecurity").text(`${resJson.linkExpiration} minute(s)`)
                for (i = 0; i < resJson.binaries.length; i++) {
                    const binary = resJson.binaries[i]
                    $( "#downloadLink" ).append(`<p><a href='https://floxy.io/api/download/${resJson.fingerPrint}/${binary.fingerPrint}/floxy'><code>${binary.kind}>./floxy -h {{host:ip}}</code>download</a> or share floxy ${binary.kind} binary (${binary.os} distro)</p>`);
                    console.log(resJson.binaries[i])
                }
                $("#sharePage").fadeIn("slow");
                return true
            })
        }
        if (resJson.status === 'expired') {
            $("#canvas").fadeTo( "slow" , 0.5)
            stopBurn();
            $("#noLinkSharePage").fadeIn("slow");
            return true
        }
        if (resJson.status === 'burning') {
            return false
        }
    }else{
        $("#burningPage").fadeOut("slow",function() {
            $("#noLinkSharePage").fadeIn("slow");
            return true
        })
    }
    return true
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
