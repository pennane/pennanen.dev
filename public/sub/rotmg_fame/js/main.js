'use strict';
$(document).ready(function () {
    $('input').on('change click paste input', function () {
        updatefame();
    });
});

function updatefame() {
    $("input.bonus-checkbox:not(:checked)").each(function () {
        $(this).parent().parent().find(':nth-child(3)').text("0");
    });
    /* Read basefame */
    var basefame, famecalc, totalfame, oldfc;
    if (isNaN(parseInt($('#equippercent').val()))) {
        $('#well-equipped').data('percentage', 0);
    } else {
        $('#well-equipped').data('percentage', parseInt($('#equippercent')
            .val()));
    }
    if (isNaN(parseInt($('#basefame').val()))) {
        basefame = 0;
    } else {
        basefame = parseInt($('#basefame').val());
    }
    famecalc = basefame;
    $(".important:checked").each(function () {});
    $('input.bonus-checkbox:checked').each(function () {
        oldfc = famecalc;
        var percent = (parseFloat($(this).parent().parent().find(
            ':first-child').data('percentage'))) + 1;
        famecalc = Math.floor((famecalc * percent).toFixed(1));
        if ($(this).parent().parent().find(':first-child').attr(
                "data-important")) {
            famecalc = famecalc + parseInt($(this).parent().parent()
                .find(':first-child').data('important'));
        }
        $(this).parent().parent().find(':nth-child(3)').text("+" + (famecalc - oldfc));
    });
    $('input.bonus-text').each(function () {
        oldfc = famecalc;
        var percent = (parseFloat($(this).parent().parent().find(
            ':first-child').data('percentage')));
        percent = (percent / 100) + 1;
        famecalc = Math.floor((famecalc * percent).toFixed(1));
        $(this).parent().parent().find(':nth-child(3)').text("+" + (famecalc - oldfc));
    });
    totalfame = parseInt(famecalc);
    $("#total-fame").text(totalfame);
    
}
