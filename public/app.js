$(function() {
	ws = new WebSocket("ws://" + location.host + "/ws");
	ws.onmessage = function(e) {
		var model = JSON.parse(e.data);
		if (model.length == 0) {
			$("#nobuilds").show()
			$("#buildlist").hide()
		} else {
			$("#nobuilds").hide()
			$("#buildlist").show()
			ret = ""
			$.each(model, function( index, value ){
				ret += "ID: " + value.ID + "<br />" + value.Start + "<br />" + value.Reference + "<br/><hr>";
			});
			$('#buildlist').html(ret)
		}
	}
	ws.onopen = function() {
	}
	ws.onclose = function() {
	}
});