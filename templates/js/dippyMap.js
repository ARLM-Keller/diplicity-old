
function dippyMap(container) {
	var that = this;
	var el = null;
	var SVG = "http://www.w3.org/2000/svg";
	if (container.find('svg').length > 0) {
	  el = container.find('svg')[0];
	}
	that.centerOf = function(province) {
		var center = $(el).find('#' + selEscape(province) + "Center").first();
		var match = /^m\s+([\d-.]+),([\d-.]+)\s+/.exec(center.attr('d'));
		var x = Number(match[1]);
		var y = Number(match[2]);
		var parentTransform = center.parent().attr("transform");
		if (parentTransform != null) {
			var transMatch = /^translate\(([\d.-]+),\s*([\d.-]+)\)$/.exec(parentTransform);
			x += Number(transMatch[1]) - 1.5;
			y += Number(transMatch[2]) - 2;
		}
		return new Poi(x,y);
	};
	that.showProvinces = function() {
		$(el).find('#provinces')[0].removeAttribute('style');
	};
	that.copySVG = function(sourceId) {
		var source = $('#' + sourceId + ' svg').first().clone();
		container[0].innerHTML = '';
		container[0].appendChild(source[0]);
		el = container.find('svg')[0];
	};
	that.colorProvince = function(province, color) {
		var path = $(el).find('#' + selEscape(province))[0];
		path.removeAttribute('style');
		path.setAttribute('fill', color);
		path.setAttribute('fill-opacity', '0.8');
	};
  that.hideProvince = function(province) {
		var path = $(el).find('#' + selEscape(province))[0];
		path.removeAttribute('style');
		path.setAttribute('fill', '#ffffff');
		path.setAttribute('fill-opacity', '0');
	};
	that.highlightProvince = function(province) {
		var prov = $(el).find('#' + selEscape(province)).first();
		var copy = prov.clone()[0];
		copy.setAttribute("id", prov.attr('id') + "_highlight");
		copy.setAttribute("style", "fill:url(#stripes)");
		copy.setAttribute('fill-opacity', '1');
		copy.removeAttribute("transform");
		var x = 0;
		var y = 0;
		var curr = prov[0];
		while (curr != null && curr.getAttribute != null) {
			var trans = curr.getAttribute("transform");
			if (trans != null) {
				var transMatch = /^translate\(([\d.-]+),\s*([\d.-]+)\)$/.exec(trans);
				x += Number(transMatch[1]);
				y += Number(transMatch[2]);
			}
			curr = curr.parentNode;
		}
		copy.setAttribute("transform", "translate(" + x + "," + y + ")");
		copy.setAttribute("stroke", 'none');
		$(el).find('#highlights')[0].appendChild(copy);
	};
	that.unhighlightProvince = function(province) {
		$(el).find('#' + selEscape(province) + '_highlight').remove();
	};
	that.addClickListener = function(province, handler, options) {
		var nohighlight = (options || {}).nohighlight;
		if (!nohighlight) {
			highlightProvince(province);
		}
		var prov = $(el).find('#' + selEscape(province)).first();
		var copy = prov.clone()[0];
		copy.setAttribute("id", prov.attr('id') + "_click");
		copy.setAttribute("style", "fill:#000000;fill-opacity:0;stroke:none;");
		copy.setAttribute("stroke", "none");
		copy.removeAttribute("transform");
		var x = 0;
		var y = 0;
		var curr = prov[0];
		while (curr != null && curr.getAttribute != null) {
			var trans = curr.getAttribute("transform");
			if (trans != null) {
				var transMatch = /^translate\(([\d.-]+),\s*([\d.-]+)\)$/.exec(trans);
				x += Number(transMatch[1]);
				y += Number(transMatch[2]);
			}
			curr = curr.parentNode;
		}
		copy.setAttribute("transform", "translate(" + x + "," + y + ")");
		el.appendChild(copy);
		var ham = $(copy).hammer({
		  prevent_default: true,
		});
		ham.bind('tap', handler);
		return function() {
			if (!nohighlight) {
				that.unhighlightProvince(province); 
			}
			ham.unbind('tap', handler);
		};
	};
	that.addBox = function(province, corners, color) {
		var loc = that.centerOf(province);
		loc.x -= 3;
		loc.y -= 3;
		var all = Math.PI * 2;
		var step = all / corners;
		var startAngle = Math.PI * 1.5;
		if (corners % 2 == 0) {
			startAngle += step / 2;
		}
		var angle = startAngle;
		var path = document.createElementNS(SVG, "path");
		path.setAttribute("style", "fill-rule:evenodd;fill:" + color + ";stroke:#000000;stroke-width:0.5;stroke-miterlimit:4;stroke-opacity:1.0;fill-opacity:0.9;");
		var d = "";
		var subBox = function(boundF) {
			d += ("M " + (loc.x + Math.cos(angle) * boundF) + "," + (loc.y + Math.sin(angle) * boundF));
			for (var i = 1; i < corners; i++) {
				angle += step;
				d += " L " + (loc.x + Math.cos(angle) * boundF) + "," + (loc.y + Math.sin(angle) * boundF);
			}      
			d += " z";
		};
    subBox(27);
		subBox(20);
		path.setAttribute("d", d);
		el.appendChild(path);
	};
	that.addArrow = function(provs, color) {
	  var start = null;
		var middle = null;
		var end = null;
		if (provs.length == 3 && provs[1] == provs[2]) {
			provs.pop();
		}
		if (provs.length == 2) {
		  start = that.centerOf(provs[0]);
			end = that.centerOf(provs[1]);
			middle = start.add(end.sub(start).div(2.0));
		} else {
		  start = that.centerOf(provs[0]);
			middle = that.centerOf(provs[1]);
			end = that.centerOf(provs[2]);
		}
		var boundF = 2;
		var headF1 = boundF * 2;
		var headF2 = boundF * 4;
		var spacer = boundF * 2;
		var boundFDiag = Math.sqrt(Math.pow(boundF, 2) + Math.pow(boundF, 2));
		var part1 = new Vec(start, middle);
		var part2 = new Vec(middle, end);
		var all = new Vec(start, end);
		var start0 = start.add(part1.dir().mul(spacer)).add(part1.orth().mul(boundF));
		var start1 = start.add(part1.dir().mul(spacer)).sub(part1.orth().mul(boundF));
		var sumOrth = part1.orth().add(part2.orth());
		var avgOrth = sumOrth.div(sumOrth.len());
		var control0 = middle.add(avgOrth.mul(boundF));
		var control1 = middle.sub(avgOrth.mul(boundF));
		var end0 = end.sub(part2.dir().mul(spacer + headF2)).add(part2.orth().mul(boundF));
		var end1 = end.sub(part2.dir().mul(spacer + headF2)).sub(part2.orth().mul(boundF));
		var end3 = end.sub(part2.dir().mul(spacer));
		var head0 = end0.add(part2.orth().mul(headF1));
		var head1 = end1.sub(part2.orth().mul(headF1));

		var path = document.createElementNS(SVG, "path");
		path.setAttribute("style", "fill:" + color + ";stroke:#000000;stroke-width:0.5;stroke-miterlimit:4;stroke-opacity:1.0;fill-opacity:0.7;");
		var d = "M " + start0.x + "," + start0.y;
		d += " C " + control0.x + "," + control0.y + "," + control0.x + "," + control0.y + "," + end0.x + "," + end0.y;
		d += " L " + head0.x + "," + head0.y;
		d += " L " + end3.x + "," + end3.y;   
		d += " L " + head1.x + "," + head1.y;
		d += " L " + end1.x + "," + end1.y;
		d += " C " + control1.x + "," + control1.y + "," + control1.x + "," + control1.y + "," + start1.x + "," + start1.y;
		d += " z";
		path.setAttribute("d", d);
		el.appendChild(path);
	};
	that.addCross = function(province, color) {
		var bound = 14;
		var width = 4;
		var loc = that.centerOf(province);
		loc.x -= 3;
		loc.y -= 3;
		var path = document.createElementNS(SVG, "path");
		path.setAttribute("style", "fill:" + color + ";stroke:#000000;stroke-width:0.5;stroke-miterlimit:4;stroke-opacity:1.0;fill-opacity:0.9;");
    var d = (
			"M " + loc.x + "," + (loc.y + width) + " " + 
			"L " + (loc.x + bound) + "," + (loc.y + bound + width) + " " + 
			"L " + (loc.x + bound + width) + "," + (loc.y + bound) + " " + 
			"L " + (loc.x + width) + "," + loc.y + " " +
			"L " + (loc.x + bound + width) + "," + (loc.y - bound) + " " +
			"L " + (loc.x + bound) + "," + (loc.y - bound - width) + " " +
			"L " + loc.x + "," + (loc.y - width) + " " +
			"L " + (loc.x - bound) + "," + (loc.y - bound - width) + " " +
			"L " + (loc.x - bound - width) + "," + (loc.y - bound) + " " +
			"L " + (loc.x - width) + "," + loc.y + " " +
			"L " + (loc.x - bound - width) + "," + (loc.y + bound) + " " +
			"L " + (loc.x - bound) + "," + (loc.y + bound + width) + " z");
		path.setAttribute("d", d);
		el.appendChild(path);
	};
	that.addOrder = function(order, variant, nation) {
	  var color = variantMap[variant].Colors[nation];
	  if (order[1] == 'Hold') {
		  addBox(order[0], 4, color);
		} else if (order[1] == 'Move') {
		  addArrow([order[0], order[2]], color);
		} else if (order[1] == 'MoveViaConvoy') {
		  addArrow([order[0], order[2]], color);
			addBox(order[0], 5, color);
		} else if (order[1] == 'Build') {
		  addUnit(variant + 'Unit' + order[2], order[0], color, false, true);
		} else if (order[1] == 'Disband') {
		  addCross(order[0], color);
		} else if (order[1] == 'Convoy') {
			addBox(order[0], 5, color);
			addArrow([order[2], order[0], order[3]], color);
		} else if (order[1] == 'Support') {
		  if (order.length == 3) {
			  addBox(order[0], 3, color);
				addArrow([order[2], order[3]], color);
			} else {
			  addBox(order[0], 3, color);
				addArrow([order[0], order[2], order[3]], color);
			}
		}
	};
	that.addUnit = function(sourceId, province, color, dislodged, build) {
		var shadow = $('#' + sourceId).find('#shadow').first().clone();
		var hullQuery = $('#' + sourceId + ' svg').find('#hull');
		var bodyQuery = $('#' + sourceId + ' svg').find('#body');
		var loc = that.centerOf(province);
		var unit = null;
		var opacity = 1;
		if (dislodged) {
			loc.x += 5;
			loc.y += 5;
			opacity = 0.73;
		}
		loc.y -= 11;
		if (hullQuery.length > 0) {
			unit = hullQuery.first().clone();
			loc.x -= 65;
			loc.y -= 15;
		} else {
			unit = bodyQuery.first().clone();
			loc.x -= 40;
			loc.y -= 5;
		}
		shadow.attr("transform", "translate(" + loc.x + ", " + loc.y + ")");
		unit.attr("transform", "translate(" + loc.x + ", " + loc.y + ")");
		if (build) {
			color = '#000000';
		}
		unit.attr("style", "fill:" + color + ";fill-opacity:" + opacity + ";stroke:#000000;stroke-width:1;stroke-miterlimit:4;stroke-opacity:1;stroke-dasharray:none");
		el.appendChild(shadow[0]);
		el.appendChild(unit[0]);
	};
	return that;
}




