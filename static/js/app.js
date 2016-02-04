$(document).ready(function() {
    $(".url-container").hover(function() {
        $(this).find(".url-actions > span").removeClass('hidden');
    },
    function() {
        $(this).find(".url-actions > span").addClass('hidden');
    });
});
