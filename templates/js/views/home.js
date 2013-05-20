window.HomeView = BaseView.extend({

	template: _.template($('#home_underscore').html()),

	initialize: function(options) {
		this.user = options.user;
	},

	render: function() {
		this.$el.html(this.template({
			user:this.user,
		}));
		new CurrentGameMembersView({ 
			el: this.$('.homePageGames'),
			collection: this.collection,
			user: this.user,
		}).doRender();
		return this;
	},

});
