const Http = new XMLHttpRequest();

$( document ).ready(function() {
    flow();
});

window.onpopstate = function(event) {
    flow();
};

function flow() {
    $("#clientServer").hide();
    $("#c2sPage").hide();
    $("#w2sPage").hide();
    $("#home").hide();
    $("#aboutPage").hide();
    $("#formPage").hide();
    $("#burnPage").hide();
    $("#noLinkSharePage").hide();
    $('#burningPage').hide();
    $('#sharePage').hide();
    if (window.location.pathname.includes('c2s')) {
        $("#clientServer").show();
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

function clientOptServer() {
    $("#home_p1").fadeOut("slow",function() {

    })
}

function goToFormPage() {
    $("#canvas").fadeTo( "slow" , 0.5)
    window.history.pushState({}, 'Form', '/form');
    $("#burnPage").fadeOut("slow",function() {
        $("#formPage").fadeIn("slow");
    })
}

function clientServer() {
    $("#canvas").fadeTo( "slow" , 1)
    window.history.pushState({}, 'Burn floxy', '/c2s');
    $("#home").fadeOut("slow",function() {
        $("#clientServer").fadeIn("slow");
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

function submit(kind) {
    grecaptcha.ready(function() {
        grecaptcha.execute('6LeUMCMaAAAAABmP3FZtGTOFcGDgpGR0Z0pI7j2R', {action: 'submit'}).then(function(token) {
            burnSubmit(token, kind)
        });
    });
}

function burnSubmit(token, kind) {
    $("#canvas").fadeTo( "slow" , 1)
    $("#home").fadeOut("slow",function() {
        startBurn()
        $("#burningPage").fadeIn("slow",function() {
            Http.open("POST", `http://localhost:8080/api/${kind}`);
            Http.setRequestHeader("Content-Type", "application/json");
            const data = JSON.stringify(
                {
                    "captcha": token,
                }
            );
            Http.send(data);
            Http.onreadystatechange = (e) => {
                const res = Http.responseText;
                if (Http.status === 201 && Http.readyState === 4) {
                    setTimeout(function(){
                        const resJson = JSON.parse(res);
                        window.history.pushState({}, 'Share floxy', `/share/${resJson.id}`);
                        getFloxy(resJson.id)
                    }, 2000);

                }
            }
        })
    });
}

function getFloxy(id) {
    // $(".copyLink").attr("href", `http://localhost:8080/api/download/${fingerprint}/floxy`)

    Http.open("GET", `http://localhost:8080/api/users/${id}`);
    Http.send();
    Http.onreadystatechange = (e) => {
        const res = Http.responseText;
        if (Http.readyState === 4) {
            if (!getFloxyResult(Http.status, res)){
                iterateUntilResult(id)
            }
        }
    }
}

function iterateUntilResult(id){
    const loop = setInterval(function(){
        Http.open("GET", `http://localhost:8080/api/users/${id}`);
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
    if (status === 404) {
        $("#noLinkSharePage").fadeIn("slow");
        return true
    }
    if (status === 200) {
        const resJson = JSON.parse(res);
        $("#burningPage").fadeOut("slow",function() {
            $("#canvas").fadeTo( "slow" , 0.5)
            greenBurn();
            $(".serverPlaceholder").text(resJson.server_command)
            $(".serverPlaceholder").text(resJson.server_command)
            $("#localPlaceholder").text(resJson.local_command)
            $(".passwordPlaceholder").text(resJson.password)
            $(".domainPlaceholder").text(resJson.domain)

            if (resJson.kind == "c2s") {
                $("#c2sPage").fadeIn("slow");
            }
            if (resJson.kind == "w2s") {
                $("#w2sPage").fadeIn("slow");
            }
            return true
        })
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

function copyText(element) {
    var $temp = $("<input>");
    $("body").append($temp);
    $temp.val($(element).text()).select();
    document.execCommand("copy");
    $temp.remove();
}

function goToHome() {
    window.history.pushState({}, 'Home floxy', `/`);
    flow()
}