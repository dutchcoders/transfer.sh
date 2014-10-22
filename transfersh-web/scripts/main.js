$(document).ready(function() {

    hljs.initHighlightingOnLoad();
    // Terminal typing animation
    /*    $("#from-terminal p").typed({
            strings: ["curl --upload-file ./hello.txt https://transfer.sh/hello.txt\n######################################################\nhttps://transfer.sh/66nb8/hello.txt \n "],
            typeSpeed: 0, // typing speed
            backSpeed: 0, // backspacing speed
            startDelay: 0, // time before typing starts
            backDelay: 500, // pause before backspacing
            loop: false, // loop on or off (true or false)
            loopCount: false, // number of loops, false = infinite
            showCursor: true,
            attr: null, // attribute to type, null = text for everything except inputs, which default to placeholder
            callback: function(){ } // call function after typing is done
        });
    */
    var typewriter = require('typewriter');

    var twSpan = document.getElementById('terminal-code');

    var tw = typewriter(twSpan).withAccuracy(100)
        .withMinimumSpeed(17)
        .withMaximumSpeed(25)
        .build();

    tw.put('$ ')
        .waitRange(500, 1000)
        .type('curl --upload-file ./hello.txt https://transfer.sh/hello.txt')
        .put('<br/>')
        .put('https://transfer.sh/66nb8/hello.txt ')
        .put('<br/>')
        .put('$ ')
        .waitRange(500, 1000)
        .put('<br/>')
        .put('$ ')
        .waitRange(500, 1000)
        .put('<br/>')
        .put('$ ')
        .waitRange(500, 1000)
        .type('transfer hello.txt')
        .put('<br/>')
        .type('####################################################')
        .put(' 100.0%')
        .put('<br/>')
        .put('https://transfer.sh/eibhM/hello.txt ')
        .put('<br/>')
        .put('$ ')
        .waitRange(1000, 1500)
        .put('<br/>')
        .put('$ ')

    // Smooth scrolling
    $('a[href*=#]:not([href=#])').click(function() {
        if (location.pathname.replace(/^\//, '') == this.pathname.replace(/^\//, '') && location.hostname == this.hostname) {
            var target = $(this.hash);
            target = target.length ? target : $('[name=' + this.hash.slice(1) + ']');
            if (target.length) {
                $('html,body').animate({
                    scrollTop: target.offset().top
                }, 1000);
                return false;
            }
        }
    });


    // function resizePages() {
    //     var h = $(window).height();
    //     var height  =  h < 600 ? 600 : h;
    /*  $('section').css('height',height);
        $('#home').css('height',height*0.98);
    }
    resizePages();*/
});

(function() {
    var files = Array()

    function upload(file) {

        $('.browse').addClass('uploading');
        var li = $('<li style="clear:both;"/>');

        li.append($('<div><div class="progress active upload-progress" style="margin-bottom: 0;"><div class="progress-bar bar" style="width: 0%;"></div></div><p>Uploading... ' + file.name + '</p></div>'));
        $(li).appendTo($('.queue'));

        var xhr = new XMLHttpRequest();

        xhr.upload.addEventListener("progress", function(e) {
            var pc = parseInt((e.loaded / e.total * 100));
            $('.upload-progress', $(li)).show();
            $('.upload-progress .bar', $(li)).css('width', pc + "%");
        }, false);

        xhr.onreadystatechange = function(e) {
            if (xhr.readyState == 4) {
                $('.upload-progress', $(li)).hide();
                $('#web').addClass('uploading');
                // progress.className = (xhr.status == 200 ? "success" : "failure");
                if (xhr.status == 200) {
                    $(li).html('<a target="_blank" href="' + xhr.responseText + '">' + xhr.responseText + '</a>');
                } else {
                    $(li).html('<span>Error (' + xhr.status + ') during upload of file ' + file.name + '</span>');
                }

                files.push(xhr.responseText.replace("https://transfer.sh/", "").replace("\n", ""));
                // files.push(URI(xhr.responseText).absoluteTo(location.href).toString());

                $(".download-zip").attr("href", URI("(" + files.join(",") + ").zip").absoluteTo(location.href).toString());
                $(".download-tar").attr("href", URI("(" + files.join(",") + ").tar.gz").absoluteTo(location.href).toString());

                $(".all-files").addClass('show');

            }

        };

        // should queue all uploads. 

        // start upload
        xhr.open("PUT", '/' + file.name, true);
        xhr.setRequestHeader("X_FILENAME", file.name);
        xhr.send(file);
    };

    $(document).bind("dragenter", function(event) {

        event.preventDefault();
    }).bind("dragover", function(event) {
        event.preventDefault();
        // show drop indicator
        $('#web').addClass('dragged');
    }).bind("dragleave", function(event) {
        $('#web').removeClass('dragged');
        console.log('asdasd');

    }).bind("drop dragdrop", function(event) {
        var files = event.originalEvent.target.files || event.originalEvent.dataTransfer.files;

        $.each(files, function(index, file) {
            console.debug(file);
            upload(file);
        });

        event.stopPropagation();
        event.preventDefault();
    });

    $('a.browse').on('click', function(event) {
        $("input[type=file]").click();
        return (false);
    });



    $('input[type=file]').on('change', function(event) {
        $.each(this.files, function(index, file) {
            if (file instanceof Blob) {
                upload(file);
            }
        });
    });
})();


