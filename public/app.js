function timeSince(date) {
	var seconds = Math.floor((new Date() - date) / 1000);

	times = [];

	interval = Math.floor(seconds / 3600);
	if (interval > 1) {
		times.push(interval + " h");
		seconds -= interval * 3600;
	}
	interval = Math.floor(seconds / 60);
	if (interval > 1) {
		times.push(interval + " min");
		seconds -= interval * 60;
	}
	times.push(Math.floor(seconds) + " sec");
	return times.join(", ");
}

$(function() {
	var wssUri = "ws://" + location.host + $('#baseUri').val() + "ws";
	if (location.protocol === 'https:') {
		wssUri = "wss://" + location.host + $('#baseUri').val() + "ws";
	}

	ws = new ReconnectingWebSocket(wssUri);
	ws.onmessage = function(e) {
		var d = JSON.parse(e.data);
		if (typeof d.running == "undefined" || d.running.length == 0) {
			$("#nobuilds").show()
			$("#buildlist").hide()
		} else {
			var ret = ""
			$.each(d.running, function( index, value ){
				tpl = $('#buildtemplate').html()
				tpl = tpl.replace(/##JOBID##/g, value.ID);
				tpl = tpl.replace(/##COMMIT##/g, value.Reference.substring(0, 7));
				tpl = tpl.replace(/##STATUS##/g, value.Status);
				tpl = tpl.replace(/##START##/g, value.Start);
				tpl = tpl.replace(/##SINCE##/g, timeSince(new Date(value.Start)));
				ret += tpl;
			});
			$('#buildlist').html(ret)
			$("#nobuilds").hide()
			$("#buildlist").show()
		}
		if (typeof d.queue == "undefined" || d.queue.length > 0) {
			var ret = ""
			$.each(d.queue, function( index, value ){
				tpl = $('#buildtemplate').html()
				tpl = tpl.replace(/##JOBID##/g, value.ID);
				tpl = tpl.replace(/##COMMIT##/g, value.Reference.substring(0, 7));
				tpl = tpl.replace(/##STATUS##/g, value.Status);
				tpl = tpl.replace(/##START##/g, value.CreatedAt);
				tpl = tpl.replace(/##SINCE##/g, timeSince(new Date(value.CreatedAt)));
				ret += tpl;
			});
			$('#buildqueue').html(ret)
		}
	}
	ws.onopen = function() {
	}
	ws.onclose = function() {
	}

	$('.build-branch').on('click', function(e) {
		e.preventDefault();
		ws.send(JSON.stringify({
			"action":"build",
			"id": parseInt($(this).attr('data-id'))
		}));
	});
	function timestampUpdater() {
		$('.time-block').each(function(i) {
			$(this).html(timeSince(new Date($(this).attr('data-id'))));
		})
		setTimeout(timestampUpdater, 1000);
	}
	setTimeout(timestampUpdater, 1000);
});