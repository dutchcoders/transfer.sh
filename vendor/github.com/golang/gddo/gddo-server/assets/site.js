// jump modal
$(function() {

    var all;
    var visible;
    var active = -1;
    var lastFilter = '';
    var $body = $('#x-jump-body');
    var $list = $('#x-jump-list');
    var $filter = $('#x-jump-filter');
    var $modal = $('#x-jump');

    var update = function(filter) {
        lastFilter = filter;
        if (active >= 0) {
            visible[active].e.removeClass('active');
            active = -1;
        }
        visible = []
        var re = new RegExp(filter.replace(/([.*+?^=!:${}()|\[\]\/\\])/g, "\\$1"), "gi");
        all.forEach(function (id) {
            id.e.detach();
            var text = id.text;
            if (filter) {
                text = id.text.replace(re, function (s) { return '<b>' + s + '</b>'; });
                if (text == id.text) {
                    return
                }
            }
            id.e.html(text + ' ' + '<i>' + id.kind + '</i>');
            visible.push(id);
        });
        $body.scrollTop(0);
        if (visible.length > 0) {
            active = 0;
            visible[active].e.addClass('active');
        }
        $list.append($.map(visible, function(identifier) { return identifier.e; }));
    }

    var incrActive = function(delta) {
        if (visible.length == 0) {
            return
        }
        visible[active].e.removeClass('active');
        active += delta;
        if (active < 0) {
            active = 0;
            $body.scrollTop(0);
        } else if (active >= visible.length) {
            active = visible.length - 1;
            $body.scrollTop($body[0].scrollHeight - $body[0].clientHeight);
        } else {
            var $e = visible[active].e;
            var t = $e.position().top;
            var b = t + $e.outerHeight(false);
            if (t <= 0) {
                $body.scrollTop($body.scrollTop() + t);
            } else if (b >= $body.outerHeight(false)) {
                $body.scrollTop($body.scrollTop() + b - $body.outerHeight(false));
            }
        }
        visible[active].e.addClass('active');
    }

    $modal.on('show.bs.modal', function() {
        if (!all) {
            all = []
            var kinds = {'c': 'constant', 'v': 'variable', 'f': 'function', 't': 'type', 'd': 'field', 'm': 'method'}
            $('*[id]').each(function() {
                var e = $(this);
                var id = e.attr('id');
                if (/^[^_][^-]*$/.test(id)) {
                    all.push({
                        text: id,
                        ltext: id.toLowerCase(),
                        kind: kinds[e.closest('[data-kind]').attr('data-kind')],
                        e: $('<a/>', {href: '#' + id, 'class': 'list-group-item', tabindex: '-1'})
                    });
                }
            });
            all.sort(function (a, b) {
                if (a.ltext > b.ltext) { return 1; }
                if (a.ltext < b.ltext) { return -1; }
                return 0
            });
        }
    }).on('shown.bs.modal', function() {
        update('');
        $filter.val('').focus();
    }).on('hide.bs.modal', function() {
        $filter.blur();
    }).on('click', '.list-group-item', function() {
        $modal.modal('hide');
    });

    $filter.on('change keyup', function() {
        var filter = $filter.val();
        if (filter.toUpperCase() != lastFilter.toUpperCase()) {
            update(filter);
        }
    }).on('keydown', function(e) {
        switch(e.which) {
        case 38: // up
            incrActive(-1);
            e.preventDefault(); 
            break;
        case 40: // down
            incrActive(1);
            e.preventDefault(); 
            break;
        case 13: // enter
            if (active >= 0) {
                visible[active].e[0].click();
            }
            break
        }
    });

});

$(function() {

    if ("onhashchange" in window) {
        var highlightedSel = "";
        window.onhashchange = function() {
            if (highlightedSel) {
                $(highlightedSel).removeClass("highlighted");
            }
            highlightedSel = window.location.hash.replace( /(:|\.|\[|\]|,)/g, "\\$1" );
            if (highlightedSel && (highlightedSel.indexOf("example-") == -1)) {
                $(highlightedSel).addClass("highlighted");
            }
        };
        window.onhashchange();
    }

});

// keyboard shortcuts
$(function() {
    var prevCh = null, prevTime = 0, modal = false;

    $('.modal').on({
        show: function() { modal = true; },
        hidden: function() { modal = false; }
    });

    $(document).on('keypress', function(e) {
        var combo = e.timeStamp - prevTime <= 1000;
        prevTime = 0;

        if (modal) {
            return true;
        }

        var t = e.target.tagName
        if (t == 'INPUT' ||
            t == 'SELECT' ||
            t == 'TEXTAREA' ) {
            return true;
        }

        if (e.target.contentEditable && e.target.contentEditable == 'true') {
            return true;
        }

        if (e.metaKey || e.ctrlKey) {
            return true;
        }

        var ch = String.fromCharCode(e.which);

        if (combo) {
            switch (prevCh + ch) {
            case "gg":
                $('html,body').animate({scrollTop: 0},'fast');
                return false;
            case "gb":
                $('html,body').animate({scrollTop: $(document).height()},'fast');
                return false;
            case "gi":
                if ($('#pkg-index').length > 0) {
                    $('html,body').animate({scrollTop: $("#pkg-index").offset().top},'fast');
                    return false;
                }
            case "ge":
                if ($('#pkg-examples').length > 0) {
                    $('html,body').animate({scrollTop: $("#pkg-examples").offset().top},'fast');
                    return false;
                }
            }
        }

        switch (ch) {
        case "/":
            $('#x-search-query').focus();
            return false;
        case "?":
            $('#x-shortcuts').modal();
            return false;
        case  "f":
            if ($('#x-jump').length > 0) {
                $('#x-jump').modal();
                return false;
            }
        }

        prevCh = ch
        prevTime = e.timeStamp
        return true;
    });
});

// misc
$(function() {
    $('span.timeago').timeago();
    if (window.location.hash.substring(0, 9) == '#example-') {
        var id = '#ex-' + window.location.hash.substring(9);
        $(id).addClass('in').removeClass('collapse').height('auto');
    }

    $(document).on("click", "input.click-select", function(e) {
        $(e.target).select();
    });

    $('body').scrollspy({
        target: '.gddo-sidebar',
        offset: 10
    });
});
