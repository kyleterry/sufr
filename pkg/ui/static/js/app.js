$(document).ready(function() {
  $('#confirm-delete').on('show.bs.modal', function(e) {
    $(this).find('.btn-ok').attr('href', $(e.relatedTarget).data('href'));
  });

  $('.fav-btn').click(function() {
    var btn = this;
    var url = btn.href;
    $.ajax({
      url: url,
      type: 'post',
      success: function(result) {
        var klass = '';
        if (result.state) {
          klass = 'fa fa-heart'
        } else {
          klass = 'fa fa-heart-o'
        }
        $(btn).children('i').removeClass().addClass(klass);
      }
    });
    return false;
  });

  var vidDefer = document.getElementsByTagName('iframe');
  for (var i=0; i<vidDefer.length; i++) {
    if(vidDefer[i].getAttribute('data-src')) {
      vidDefer[i].setAttribute('src',vidDefer[i].getAttribute('data-src'));
    }
  }
});
