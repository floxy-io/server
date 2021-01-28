
$( document ).ready(function() {
    $("#home").hide();
    $("#burn").hide();
    if (window.location.pathname.includes('burn')){
        $("#burn").show();
    }else {
        $("#home").show();
    }
});

function toBurnpage() {
    window.history.pushState({}, 'Burn floxy', '/burn');
    $("#home").fadeOut("slow",function() {
        $("#burn").fadeIn("slow");
        $("#burn").css("display", "flex");
    })
}