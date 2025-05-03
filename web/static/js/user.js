console.log("loaded this guy")
$(document).ready(function () {
    $('.btn-logout').click(function (e) {
        Cookies.remove('auth-session');
    });
});