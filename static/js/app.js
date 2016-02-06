$(document).ready(function() {
    $(".url-container").hover(function() {
      $(this).find(".url-actions > span").removeClass('hidden');
    },
    function() {
      $(this).find(".url-actions > span").addClass('hidden');
    });

    $('#confirm-delete').on('show.bs.modal', function(e) {
      $(this).find('.btn-ok').attr('href', $(e.relatedTarget).data('href'));
    });
});
