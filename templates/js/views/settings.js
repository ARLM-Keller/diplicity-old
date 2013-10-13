window.SettingsView = BaseView.extend({

  template: _.template($('#settings_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.listenTo(window.session.user, 'change', this.doRender);
	},

  render: function() {
		var that = this;
		that.$el.html(that.template({
		  user: window.session.user,
		}));
		navLinks(mainButtons);
		return that;
	},

});
