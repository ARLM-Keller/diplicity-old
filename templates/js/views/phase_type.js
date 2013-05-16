window.PhaseTypeView = Backbone.View.extend({

  template: _.template($('#phase_type_underscore').html()),

  me: new Date().getTime(),

	events: {
		"change .deadline": "changeDeadline",
		"change .chat-flag": "changeChatFlag",
	},

	initialize: function(options) {
	  _.bindAll(this, 'render', 'update', 'onClose');
		this.phaseType = options.phaseType;
		this.owner = options.owner;
		this.gameMember = options.gameMember;
		this.gameMember.bind('change', this.update);
		options.parent.children.push(this);
	},

	onClose: function() {
	  this.gameMember.unbind('change', this.update);
	},

	changeDeadline: function(ev) {
		this.gameMember.get('game').deadlines[this.phaseType] = parseInt($(ev.target).val()); 
		this.gameMember.trigger('change');
		this.gameMember.trigger('saveme');
	},

  update: function() {
	  var that = this;
		var desc = [];
		for (var i = 0; i < deadlineOptions.length; i++) { 
		  var opt = deadlineOptions[i];
		  if (opt.value == that.gameMember.get('game').deadlines[that.phaseType]) {
			  desc.push(opt.name);
				that.$('.deadline').val('' + opt.value);
			}
		} 
		for (var i = 0; i < chatFlagOptions().length; i++) {
			var opt = chatFlagOptions()[i];
			if ((opt.id & that.gameMember.get('game').chat_flags[that.phaseType]) != 0) {
			  desc.push(opt.name);
				that.$('input[type=checkbox][data-chat-flag=' + opt.id + ']').attr('checked', 'checked');
			} else {
				that.$('input[type=checkbox][data-chat-flag=' + opt.id + ']').removeAttr('checked');
			}
			that.$('input[type=checkbox][data-chat-flag=' + opt.id + ']').checkboxradio().checkboxradio('refresh');
		}
		that.$('.desc').text(desc.join(", "));
		that.$('select.deadline').val(that.gameMember.get('game').deadlines[that.phaseType]);
		that.$('select.deadline').selectmenu().selectmenu('refresh');
	},

	changeChatFlag: function(ev) {
	  if ($(ev.target).is(":checked")) {
			this.gameMember.get('game').chat_flags[this.phaseType] |= parseInt($(ev.target).attr('data-chat-flag'));
		} else {
			this.gameMember.get('game').chat_flags[this.phaseType] = this.gameMember.get('game').chat_flags[this.phaseType] & (~parseInt($(ev.target).attr('data-chat-flag')));
		}
		this.gameMember.trigger('change');
		this.gameMember.trigger('saveme');
	},

  render: function() {
		this.$el.html(this.template({
		  owner: this.owner,
		  me: this.me,
		  phaseType: this.phaseType,
		}));
		this.update();
		this.$el.trigger('create');
		return this;
	},

});