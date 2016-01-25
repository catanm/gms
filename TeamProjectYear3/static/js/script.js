// $(function(){
// 	// body loaded

// 	$(addGallery());

// 	function addGallery(){
// 		var current_li;
// 		$("#portfolio img").click(function(){
// 			console.log("clicked");
// 			var src = $(this).attr("src");
// 			current_li = $(this).parent();
// 			$("#main-img").attr("ng-src", src); 
// 			$("#frame").fadeIn();
// 			$("#overlay").fadeIn(); 
// 		});

// 		$("#overlay").click(function(){
// 			$(this).fadeOut();
// 			$("#frame").fadeOut();
// 		});

// 		$("#right-arrow").click(function(){

// 			if(current_li.is(":last-child")){
// 				var next_li = $("#portfolio li").first();
// 			}else{
// 				var next_li = current_li.next();
// 			}

// 			var next_src = next_li.children("img").attr("ng-src");
// 			$("#main-img").attr("ng-src", next_src);
// 			current_li = next_li;
// 		});

// 		$("#left-arrow").click(function(){

// 			if(current_li.is(":first-child")){
// 				var prev_li = $("#portfolio li").last();
// 			}else{
// 				var prev_li = current_li.prev();
// 			}

// 			var prev_src = prev_li.children("img").attr("ng-src");
// 			$("#main-img").attr("ng-src", prev_src);
// 			current_li = prev_li;
// 		});

// 		$("#left-arrow, #right-arrow").mouseover(function(){
// 			$(this).css("opacity", "0.8");
// 		});

// 		$("#left-arrow, #right-arrow").mouseleave(function(){
// 			$(this).css("opacity", "0.6");
// 		});
// 	}
	

// });