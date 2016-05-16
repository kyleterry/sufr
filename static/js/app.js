function isBreakpoint( alias ) {
  return $('.device-' + alias).is(':visible');
}

function bindHideAction() {
  $(".url-actions > span").addClass('hidden');
  $(".url-container").hover(function() {
    $(this).find(".url-actions > span").removeClass('hidden');
  },
  function() {
    $(this).find(".url-actions > span").addClass('hidden');
  });
}

function unbindHideAction() {
  $(".url-actions > span").removeClass('hidden');
  $(".url-container").unbind('mouseenter mouseleave');
}

$(document).ready(function() {
  if( isBreakpoint('md') || isBreakpoint('lg') ) {
    bindHideAction();
  }

  // Some of this viewport detection code comes from: http://stackoverflow.com/a/22885503/1454445
  var waitForFinalEvent=function(){var b={};return function(c,d,a){a||(a="I am a banana!");b[a]&&clearTimeout(b[a]);b[a]=setTimeout(c,d)}}();
  var fullDateString = new Date();
  $(window).resize(function() {
    waitForFinalEvent(function(){
      if (isBreakpoint('xs') || isBreakpoint('sm')) {
        unbindHideAction();
      } else if (isBreakpoint('md') || isBreakpoint('lg')) {
        bindHideAction();
      }
    }, 300, fullDateString.getTime())
  });

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
});
