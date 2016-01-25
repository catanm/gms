// 'use strict';

var app = angular.module("mainModule", ["ui.router", "ui.bootstrap.datetimepicker", "mgcrea.ngStrap", "angularFileUpload", "ngAnimate", "ngCookies"]);

app.directive('maphome', ["$http",
    function($http) {
        return {
            restrict: 'AE',
            replace: true,
            link: function(scope, elem, attrs) {

                scope.$watch(function() {
                    return scope.loggedUser;
                }, function(newUser) {

                    if (newUser) {
                        var params = {};
                        params.heatmap = true;
                        params.personal = true;
                    } else {
                        var params = {};
                        params.heatmap = true;
                        params.personal = false;
                    }

                    $http.get("/api/trails/get", {
                        params: params
                    }).success(function(response) {
                        var newPoints = response.data;
                        $http.get("/api/popLocations", {
                            params: params
                        }).success(function(res) {

                            scope.width = elem[0].offsetWidth - 30;
                            scope.height = scope.width;
                            var d = res.data;
                            console.log(d);
                            var popularLocations = d.marker;
                            var recommend = d.recommend;

                            function initialise(position) {
                                //if no map container, create it.
                                if (document.getElementById('mapcontainer') == null) {
                                    scope.mapcanvas = document.createElement('div');
                                    scope.mapcanvas.id = 'mapcontainer';
                                    scope.mapcanvas.style.height = scope.height + 'px';
                                    scope.mapcanvas.style.width = scope.width + 'px';
                                    document.querySelector('#map_article').appendChild(scope.mapcanvas);
                                    scope.glasgow = new google.maps.LatLng(55.864237, -4.251806);
                                    scope.options = {
                                        zoom: 12,
                                        center: scope.glasgow
                                    }
                                }
                                //add map
                                var marker;
                                var imageUrl;
                                var urlI;
                                map = new google.maps.Map(document.getElementById('mapcontainer'), scope.options);
                                var infowindow = new google.maps.InfoWindow({
                                    disableAutoPan: true
                                });
                                var googlePoints = [];
                                var globalIndex = 0;
                                var street;

                                for (i = 0; i < newPoints.length; i++) {
                                    googlePoints.push(new google.maps.LatLng(parseFloat(newPoints[i][0]), parseFloat(newPoints[i][1])));
                                }
                                var pointArray = new google.maps.MVCArray(googlePoints);
                                heatmap = new google.maps.visualization.HeatmapLayer({
                                    data: pointArray,
                                    maxIntensity: newPoints.length / 100, //adjust intensity according to number of points 
                                });
                                heatmap.setMap(map);
                                var gradient = [
                                    'rgba(0, 255, 255, 0)',
                                    'rgba(0, 255, 255, 1)',
                                    'rgba(0, 191, 255, 1)',
                                    'rgba(0, 127, 255, 1)',
                                    'rgba(0, 63, 255, 1)',
                                    'rgba(0, 0, 255, 1)',
                                    'rgba(0, 0, 223, 1)',
                                    'rgba(0, 0, 191, 1)',
                                    'rgba(0, 0, 159, 1)',
                                    'rgba(0, 0, 127, 1)',
                                    'rgba(63, 0, 91, 1)',
                                    'rgba(127, 0, 63, 1)',
                                    'rgba(191, 0, 31, 1)',
                                    'rgba(255, 0, 0, 1)'
                                ]
                                heatmap.set('gradient', heatmap.get('gradient') ? null : gradient);
                                google.maps.event.trigger(map, 'resize');

                                for (i = 0; i < (popularLocations ? popularLocations.length : 0); i++) {

                                    if (params.personal == true) {
                                        street = popularLocations[i]["street"];
                                    } else {
                                        street = popularLocations[i]["streetName"];
                                    }
                                    var pinImage = new google.maps.MarkerImage("static/img/popular.png");
                                    marker = new google.maps.Marker({
                                        position: new google.maps.LatLng(parseFloat(popularLocations[i]["lat"]), parseFloat(popularLocations[i]["lon"])),
                                        map: map,
                                        icon: pinImage,
                                        title: "POPULAR"
                                    });

                                    if (params.personal == false) {
                                        imageUrl = popularLocations[i]["url"];

                                        google.maps.event.addListener(marker, 'mouseover', (function(marker, globalIndex, imageUrl, street) {
                                            return function() {
                                                if (imageUrl != "unknown") {
                                                    infowindow.setContent('<IMG BORDER="0" WIDTH="400" ALIGN="Left" SRC="' + imageUrl + '">' + '<br>' + '<font size="3" color="black"> <p>' + street + '</p></font>');
                                                } else {
                                                    infowindow.setContent('<font size="3" color="black"> <h4>Image Unavailable</h4> <p>' +
                                                        street + '</p></font>');
                                                }
                                                infowindow.open(map, marker);
                                            }
                                        }(marker, globalIndex, imageUrl, street)));

                                    } else {
                                        google.maps.event.addListener(marker, 'mouseover', (function(marker, globalIndex, street) {
                                            return function() {
                                                infowindow.setContent('<font size="3" color="black"> <p>' + street + '</p></font>');
                                                infowindow.open(map, marker);
                                            }
                                        }(marker, globalIndex, street)));
                                    }

                                    globalIndex++;

                                    google.maps.event.addListener(marker, 'mouseout', function() {
                                        infowindow.close();
                                    });
                                }

                                for (i = 0; i < (recommend ? recommend.length : 0); i++) {

                                    var street = recommend[i]["placeName"];
                                    var placeCategory = recommend[i]["placeCategory"];
                                    var popularity = recommend[i]["popularity"];

                                    var pinImage = new google.maps.MarkerImage("static/img/recommended.png");
                                    marker = new google.maps.Marker({
                                        position: new google.maps.LatLng(parseFloat(recommend[i]["lat"]), parseFloat(recommend[i]["lon"])),
                                        map: map,
                                        title: "RECOMMENDED",
                                        icon: pinImage
                                    });

                                    google.maps.event.addListener(marker, 'mouseover', (function(marker, globalIndex, street, placeCategory, popularity) {
                                        return function() {

                                            infowindow.setContent('<font size="3" color="black"> <p> <b>Place:</b> ' + street + '<br> <b>Place Category:</b> ' + placeCategory + '</p></font>');

                                            infowindow.open(map, marker);

                                        }
                                    }(marker, globalIndex, street, placeCategory, popularity)));

                                    globalIndex++;

                                    google.maps.event.addListener(marker, 'mouseout', function() {
                                        infowindow.close();
                                    });
                                }

                            }
                            if (navigator.geolocation) {
                                navigator.geolocation.getCurrentPosition(initialise);
                            } else {
                                error('Geo Location is not supported');
                            }
                        }).catch(function(res) {
                        });
                    }).catch(function(response) {
                    });
                });
            },
            templateUrl: 'static/templates/homeMap.html'
        };
    }
]);

app.directive('imagemap', ["$http", "$filter",
    function($http, $filter) {
        return {
            restrict: 'AE',
            replace: true,
            link: function(scope, elem, attrs) {

                scope.$watch(function() {
                    return scope.images;
                }, function(newImages) {

                    scope.width = elem[0].offsetWidth - 30;
                    scope.height = scope.width;

                    function initialise(position) {
                        //if no map container, create it.
                        if (document.getElementById('mapcontainer') == null) {
                            scope.mapcanvas = document.createElement('div');
                            scope.mapcanvas.id = 'mapcontainer';
                            scope.mapcanvas.style.height = scope.height + 'px';
                            scope.mapcanvas.style.width = scope.width + 'px';
                            document.querySelector('#map_article').appendChild(scope.mapcanvas);
                            scope.glasgow = new google.maps.LatLng(55.864237, -4.251806);
                            scope.options = {
                                zoom: 12,
                                center: scope.glasgow
                            }
                        }
                        map = new google.maps.Map(document.getElementById('mapcontainer'), scope.options);
                        var infowindow = new google.maps.InfoWindow({
                            disableAutoPan: true
                        });
                        var marker;

                        var colours = ['009933', '0099ff', 'cc33ff', 'ff0066', '66ccff', 'ffff00']

                        var globalIndex = 0;

                        newImages.keys.forEach(function(date, dateIndex) {
                            newImages[date].forEach(function(image) {

                                if (!image.lat || !image.lon) {
                                    return;
                                }

                                var pinImage = new google.maps.MarkerImage("http://www.googlemapsmarkers.com/v1/" + colours[dateIndex % colours.length] + "/");
                                marker = new google.maps.Marker({
                                    position: new google.maps.LatLng(image.lat, image.lon),
                                    map: map,
                                    icon: pinImage,
                                    title: image.date
                                });

                                google.maps.event.addListener(marker, 'mouseover', (function(marker, globalIndex) {
                                    return function() {
                                        infowindow.setContent('<IMG BORDER="0" WIDTH="500" ALIGN="Left" SRC="' + image.url + '">' +
                                            '<br>' + $filter('date')(image.date, 'd MMMM yyyy on EEEE @ HH:mm'));
                                        infowindow.open(map, marker);
                                        infowindow.setOptions()
                                    }
                                })(marker, globalIndex));

                                google.maps.event.addListener(marker, 'mouseout', function() {
                                    infowindow.close();
                                });

                                marker.setMap(map)
                                globalIndex++;
                            })
                        })
                    }
                    if (navigator.geolocation) {
                        navigator.geolocation.getCurrentPosition(initialise);
                    } else {
                        error('Geo Location is not supported');
                    }
                });
            },
            templateUrl: 'static/templates/homeMap.html'
        };
    }
]);

app.directive('map', ["$http", "$cookies",
    function($http, $cookies) {
        return {
            restrict: 'AE',
            replace: true,
            link: function(scope, elem, attrs) {

                scope.$watch(function() {
                    return scope.changeInDateOrMap;
                }, function(newDate) {

                    if (scope.mapType === "heatmap") {
                        var params = angular.copy(scope.selectedDate);
                        params.heatmap = true;
                        params.personal = scope.personal;
                        $http.get("/api/trails/get", {
                            params: params
                        }).success(function(response) {
                            personalHeatmap = response.data;
                            if (navigator.geolocation) {
                                navigator.geolocation.getCurrentPosition(initialise);
                            } else {
                                error('Geo Location is not supported');
                            }
                        }).catch(function(response) {
                            console.log(response)
                        });
                    }
                    scope.width = elem[0].offsetWidth - 30;

                    scope.height = scope.width;

                    function initialise(position) {

                        if (document.getElementById('mapcontainer') == null) {
                            scope.mapcanvas = document.createElement('div');
                            scope.mapcanvas.id = 'mapcontainer';
                            scope.mapcanvas.style.height = scope.height + 'px';
                            scope.mapcanvas.style.width = scope.width + 'px';
                            document.querySelector('#map_article').appendChild(scope.mapcanvas);
                            scope.glasgow = new google.maps.LatLng(55.864237, -4.251806);
                            scope.options = {
                                zoom: 12,
                                center: scope.glasgow
                            }
                        }

                        scope.map = new google.maps.Map(document.getElementById("mapcontainer"), scope.options);

                        if (scope.mapType === "heatmap") {
                            var googlePoints = [];
                            for (i = 0; i < personalHeatmap.length; i++) {
                                googlePoints.push(new google.maps.LatLng(parseFloat(personalHeatmap[i][0]), parseFloat(personalHeatmap[i][1])));
                            }
                            var pointArray = new google.maps.MVCArray(googlePoints);
                            heatmap = new google.maps.visualization.HeatmapLayer({
                                data: pointArray,
                                maxIntensity: personalHeatmap.length / 100
                            });
                            heatmap.setMap(scope.map);
                            var gradient = [
                                'rgba(0, 255, 255, 0)',
                                'rgba(0, 255, 255, 1)',
                                'rgba(0, 191, 255, 1)',
                                'rgba(0, 127, 255, 1)',
                                'rgba(0, 63, 255, 1)',
                                'rgba(0, 0, 255, 1)',
                                'rgba(0, 0, 223, 1)',
                                'rgba(0, 0, 191, 1)',
                                'rgba(0, 0, 159, 1)',
                                'rgba(0, 0, 127, 1)',
                                'rgba(63, 0, 91, 1)',
                                'rgba(127, 0, 63, 1)',
                                'rgba(191, 0, 31, 1)',
                                'rgba(255, 0, 0, 1)'
                            ]
                            heatmap.set('gradient', heatmap.get('gradient') ? null : gradient);
                        } else if (scope.mapType === "kmlfile") {
                            var params = angular.copy(scope.selectedDate) || {};
                            params.personal = Boolean(scope.personal);
                            url = attachParam("http://mirtest.dcs.gla.ac.uk/api/trails/get", angular.extend(params, {
                                "token": $cookies.token,
                                "ra": Math.random()
                            }))
                            console.log(url);
                            var trackLayer = new google.maps.KmlLayer({
                                url: url
                            });
                            trackLayer.setMap(scope.map);
                        }
                        //google.maps.event.addDomListener(window, 'load', initialise);
                    }
                    if (navigator.geolocation) {
                        navigator.geolocation.getCurrentPosition(initialise);
                    } else {
                        error('Geo Location is not supported');
                    }
                });
            },
            templateUrl: 'static/templates/map.html'
        };
    }
]);

app.directive("imageViewer", function() {
    return {
        restrict: "A",
        link: function(scope, elem, attrs) {
            if (scope.$last) {

                $(".portfolio img").click(function() {
                    $('#viewerModal').modal('show');
                    $(".slider").fadeIn();
                });
            }
        }
    };
});

app.directive("datepicker", function() {
    return {
        restrict: "AE",
        replace: true,
        templateUrl: 'static/templates/datepicker.html'
    }
})

app.directive("imagecountbarchart", function() {
    return {
        restrict: "E",
        link: function(scope, elem, attrs) {

            var newStats = {};

            scope.$watch(function() {
                return JSON.stringify(scope.stats); //+ JSON.stringify(scope.imagesBarchart); //{stats: scope.stats, imagesBarchart: scope.imagesBarchart};
            }, function(str) {
                newStats.stats = scope.stats;
                newStats.barchart = scope.imagesBarchart;

                if (newStats.stats.length === 0)
                    return

                var imagesData = {
                    labels: newStats.stats[1],
                    datasets: [{
                        fillColor: "rgba(151,187,205,0.5)",
                        strokeColor: "rgba(151,187,205,0.8)",
                        highlightFill: "rgba(151,187,205,0.75)",
                        highlightStroke: "rgba(151,187,205,1)",
                        data: newStats.stats[0]
                    }]
                }

                var canvas = document.getElementById("imagesNumber");
                var imagesNumber = canvas.getContext("2d");

                canvas.style.width = '100%';
                canvas.style.height = '100%';
                canvas.width = canvas.offsetWidth;
                canvas.height = canvas.offsetHeight - canvas.offsetHeight / 50;

                if (newStats.barchart != null) {
                    newStats.barchart.destroy();
                }

                scope.imagesBarchart = new Chart(imagesNumber).Line(imagesData, {
                    datasetFill: true
                });
            });
        }
    }
});

app.directive("gpscountbarchart", function() {
    return {
        restrict: "E",
        link: function(scope, elem, attrs) {

            var newStats = {};

            scope.$watch(function() {
                return JSON.stringify(scope.stats); //+ JSON.stringify(scope.imagesBarchart); //{stats: scope.stats, imagesBarchart: scope.imagesBarchart};
            }, function(str2) {

                newStats.stats = scope.stats;
                newStats.barchart = scope.gpsBarchart;

                if (newStats.stats.length === 0)
                    return

                var gpsData = {
                    labels: newStats.stats[1],
                    datasets: [{
                        fillColor: "rgba(151,187,205,0.5)",
                        strokeColor: "rgba(151,187,205,0.8)",
                        highlightFill: "rgba(151,187,205,0.75)",
                        highlightStroke: "rgba(151,187,205,1)",
                        data: newStats.stats[2]
                    }]
                }

                var canvas = document.getElementById("gpsNumber");
                var gpsNumber = canvas.getContext("2d");

                canvas.style.width = '100%';
                canvas.style.height = '100%';
                canvas.width = canvas.offsetWidth;
                canvas.height = canvas.offsetHeight - canvas.offsetHeight / 50;

                if (newStats.barchart != null) {
                    newStats.barchart.destroy();
                }

                scope.gpsBarchart = new Chart(gpsNumber).Bar(gpsData);
            });
        }
    }
});

app.factory("Page", function() {
    var data = {
        imagePageLength: 10,
        videoPageLength: 3,
        gpsPageLength: 10,
        image_gpsPageLength: 9999999,

        currentPage: 0,
        pageType: "image",
        hasNext: true
    };

    return {
        nextPage: function() {
            if (data.hasNext)
                data.currentPage++;
        },
        previousPage: function() {
            if (data.currentPage > 0) {
                data.currentPage--;
                data.hasNext = true;
            }
        },
        getCurrentPage: function() {
            return data.currentPage;
        },
        setType: function(type) {
            data.pageType = type;
        },
        getType: function() {
            return data.pageType;
        },
        getLimit: function() {
            return data[data.pageType + "PageLength"];
        },
        getSkip: function() {
            return data[data.pageType + "PageLength"] * data.currentPage;
        },
        noNext: function() {
            data.hasNext = false;
        },
        hasNext: function() {
            return data.hasNext;
        },
        hasPrevious: function() {
            return data.currentPage > 0;
        },
        reset: function(type) {
            data.imagePageLength = 10;
            data.gpsPageLength = 10;

            data.currentPage = 0;
            data.pageType = type || "image";
            data.hasNext = true;
        }
    }
});


app.controller("MainController", ["$location", "$scope", "$http", "$state", "$cookies", "Page", "$timeout",
    function($location, $scope, $http, $state, $cookies, Page, $timeout) {
        $scope.requestCount = 0;

        $scope.loggedUser = "";
        $scope.images = [];
        $scope.stats = [];
        $scope.imagesBarchart = null;
        $scope.gpsBarchart = null;
        $scope.videos = [];
        $scope.changeInDateOrMap = "";
        $scope.requestCount = 0;
        $scope.mapType = "heatmap";
        $scope.selectedDate = {};
        $scope.heatmapPoints = {};
        $scope.personal = false;
        $scope.heatmap = true;
        $scope.keyMoments = false;
        $scope.allImages = false;
        $scope.alert = {
            type: "",
            message: "",
        }

        $scope.newAlert = function(type, message) {
            $scope.alert.type = type;
            $scope.alert.message = message;
            $timeout(function() {
                $scope.alert.message = "";
            }, 2000);
        };

        $scope.errorCheck = function(response, skip) {
            if (response.errors) {
                if (!skip)
                    $scope.newAlert("error", response.errors[0]);
                return true;
            } else {
                return false;
            }
        };

        $scope.reloadRoute = function() {
            $state.reload();
        };

        $scope.reloadPage = function() {
            window.location.reload();
        }

        $scope.updateSelectedDate = function(from_date, to_date, from_time, to_time) {


            if (typeof from_date === "object") {
                $scope.selectedDate.from_day = from_date.getDate();
                $scope.selectedDate.from_month = from_date.getMonth() + 1;
                $scope.selectedDate.from_year = from_date.getFullYear();
            }
            if (typeof to_date === "object") {
                $scope.selectedDate.to_day = to_date.getDate();
                $scope.selectedDate.to_month = to_date.getMonth() + 1;
                $scope.selectedDate.to_year = to_date.getFullYear();
            }
            if (typeof from_time === "number") {
                var hours = Math.floor(from_time / 1000 / 60 / 60);
                $scope.selectedDate.from_hour = hours;
                $scope.selectedDate.from_minute = Math.floor((from_time - hours * 60 * 60 * 1000) / 1000 / 60);
            }
            if (typeof to_time === "number") {
                var hours = Math.floor(to_time / 1000 / 60 / 60);
                $scope.selectedDate.to_hour = hours;
                $scope.selectedDate.to_minute = Math.floor((to_time - hours * 60 * 60 * 1000) / 1000 / 60);
            }

            //string being watched by the heatmap/kml directive so that it can refresh points being displayed
            //when new data range is selected
            $scope.changeInDateOrMap = "" + $scope.selectedDate.from_day + $scope.selectedDate.from_month +
                $scope.selectedDate.from_year + $scope.selectedDate.to_day + $scope.selectedDate.to_month +
                $scope.selectedDate.to_year + $scope.selectedDate.from_hour + $scope.selectedDate.to_hour +
                $scope.mapType;

            if (Page.getType() === "stats") {
                $scope.getStats();
            }
            if (Page.getType() === "pstats") {
                $scope.getStats();
            }
            if (Page.getType() === "image") {
                $scope.getImages();
                console.log("called get images");

            }
            if (Page.getType() === "image_gps") {
                $scope.getImagesMap();
            }

            if (Page.getType() === "videos") {
                $scope.getVideos();
            }
        }

        $scope.getImages = function() {
            var call;
            var oldType = Page.getType();
            Page.setType("image");
            if ($scope.loggedUser) {
                console.log("images", Page.getType())
                var params = angular.extend(angular.copy($scope.selectedDate), {
                    "limit": Page.getLimit(),
                    "skip": Page.getSkip()
                });
                params.personal = true;
                if ($scope.allImages) {
                    params.allImages = true;
                }
                if ($scope.keyMoments) {
                    params.keyMoments = true;
                }
                call = $http.get("/api/images/get", {
                    params: params
                });
            } else {
                call = $http.get("/api/images/get", {
                    params: {
                        "limit": 10
                    }
                });
            }
            Page.setType(oldType);
            call.success(function(response) {
                if (response.data.length < Page.getLimit())
                    Page.noNext();
                $scope.images = parseImages(response.data);
            }).error(function(response) {});
        };


        $scope.getImagesMap = function() {
            var call;
            var oldType = Page.getType();
            Page.setType("image_gps");
            if ($scope.loggedUser) {
                var params = angular.extend(angular.copy($scope.selectedDate), {
                    "limit": Page.getLimit(),
                    "skip": Page.getSkip()
                });
                params.personal = true;
                call = $http.get("/api/images/get", {
                    params: params
                });
            } else {
                call = $http.get("/api/images/get", {
                    params: {
                        "limit": 10
                    }
                });
            }
            Page.setType(oldType);
            call.success(function(response) {
                if (response.data.length < Page.getLimit())
                    Page.noNext();
                $scope.images = parseImages(response.data);
            }).error(function(response) {});
        };

        $scope.getVideos = function() {
            var call;
            var oldType = Page.getType();
            Page.setType("video");
            if ($scope.loggedUser) {
                var params = angular.extend(angular.copy($scope.selectedDate), {
                    "limit": Page.getLimit(),
                    "skip": Page.getSkip()
                });
                params.personal = $scope.personal;
                call = $http.get("/api/videos/get", {
                    params: params
                });
            } else {
                call = $http.get("/api/videos/get", {
                    params: {
                        "limit": 3
                    }
                });
            }
            Page.setType(oldType);
            call.success(function(response) {
                if (response.data.length < Page.getLimit())
                    Page.noNext();
                $scope.videos = parseImages(response.data);
            }).error(function(response) {
            });
        };

        $scope.getHeatmap = function() {
            var call;
            if ($scope.loggedUser) {
                var params = angular.copy($scope.selectedDate);
                params.heatmap = true;
                params.personal = $scope.personal;
                call = $http.get("/api/trails/get", {
                    params: params
                });
            } else {
                call = $http.get("/api/trails/get");
            }
            call.success(function(response) {
                $scope.heatmapPoints = response.data;
            }).error(function(response) {});
        };

        $scope.getStats = function() {
            var params = angular.copy($scope.selectedDate);
            params.page = Page.getType();
            $http.get("/api/stats/get", {
                params: params
            }).success(function(response) {
                $scope.stats = response.data;
            }).error(function(response) {
            });

        };

        $scope.setMapType = function(mapType) {
            $scope.mapType = mapType;
            $scope.changeInDateOrMap = "" + $scope.selectedDate.from_day + $scope.selectedDate.from_month +
                $scope.selectedDate.from_year + $scope.selectedDate.to_day + $scope.selectedDate.to_month +
                $scope.selectedDate.to_year + $scope.selectedDate.from_hour + $scope.selectedDate.to_hour +
                $scope.mapType;
        }

        $scope.getUserInfo = function() {
            $http.get("/api/user")
                .success(function(response) {
                    if (!$scope.errorCheck(response, true)) {
                        $scope.loggedUser = response.data;
                        $scope.getImages();
                        $scope.getHeatmap();
                        $scope.getVideos();
                        $("#loginModal").modal("hide");
                    } else {
                        $scope.loggedUser = "";
                    }
                }).error(function(response) {
                    $scope.loggedUser = "";
                });
        }

        $scope.getUserInfo();

        $scope.login = function(username, password) {
            $http.post("/api/auth/local/login", {}, {
                "headers": {
                    "Authorization": "Basic " + btoa(username + ":" + password)
                }
            }).success(function(response) {
                $cookies.token = response.data;
                if (!$scope.errorCheck(response)) {
                    $scope.getUserInfo();
                } else {
                    $scope.loggedUser = "";
                }
            })
        };

        $scope.register = function(username, email, password) {
            $http.post("/api/auth/local/register", {}, {
                "headers": {
                    "Authorization": "Basic " + btoa(username + ":" + password + ":" + email)
                }
            }).success(function(response) {
                $cookies.token = response.data;
                if (!$scope.errorCheck(response)) {
                    $scope.getUserInfo();
                } else {
                    $scope.loggedUser = "";
                }
            });
        };

        $scope.logout = function() {
            $cookies.token = "";
            $scope.loggedUser = "";
            $state.go("main");
            $scope.personal = false;
            $scope.getImages();
            $scope.getHeatmap();
            $scope.getVideos();
        };

        $scope.tab = "images";
        $scope.changeTab = function(tab) {
            $scope.tab = tab;
        };

        $scope.tooltip = {
            "title": "Hello Tooltip<br />This is a multiline message!",
            "checked": false
        };

        $scope.setPersonal = function(bool) {
            $scope.personal = bool;
        }

        $scope.setKeyMoments = function(bool) {
            $scope.keyMoments = bool;
        }

        $scope.setAllImages = function(bool) {
            $scope.allImages = bool;
        }

        $scope.clearDates = function() {
            $scope.selectedDate = null;
        };


        $scope.updateVisible = function(oldIndex, newIndex) {
            if ($scope.previewImages.length > oldIndex)
                $scope.previewImages[oldIndex].visible = false;
            if ($scope.previewImages && $scope.previewImages.length > newIndex)
                $scope.previewImages[newIndex].visible = true;
        }

        $scope.getImages();
        $scope.getVideos();
        $scope.getHeatmap();

        $scope.previewImages = [];
        $scope.currentIndex = 0;

        $scope.next = function() {
            var oldIndex = $scope.currentIndex;
            if ($scope.currentIndex < $scope.previewImages.length - 1) {
                $scope.currentIndex += 1;
            } else {
                $scope.currentIndex = 0;
            }
            $scope.updateVisible(oldIndex, $scope.currentIndex);
        };

        $scope.prev = function() {
            var oldIndex = $scope.currentIndex;
            if ($scope.currentIndex > 0) {
                $scope.currentIndex -= 1;
            } else {
                $scope.currentIndex = $scope.previewImages.length - 1;
            }
            $scope.updateVisible(oldIndex, $scope.currentIndex);
        };

        $scope.show_picture = function(index, images) {
            if (Array.isArray(images)) {
                $scope.previewImages = images;
                if ($scope.previewImages && $scope.previewImages.length > $scope.currentIndex)
                    $scope.previewImages[$scope.currentIndex].visible = false;
                $scope.currentIndex = index;
                if ($scope.previewImages && $scope.previewImages.length > $scope.currentIndex)
                    $scope.previewImages[$scope.currentIndex].visible = true;
            }
        };
    }
]);

app.controller("master", ["$rootScope",
    function($rootScope) {
        $rootScope.$on('$stateChangeError', function(event, toState, toParams, fromState, fromParams, error) {
            console.log("Routing error");
            console.log(event);
            console.log(toState);
            console.log(toParams);
            console.log(fromState);
            console.log(fromParams);
        });
    }
]);

app.config(["$stateProvider", "$urlRouterProvider",
    function($stateProvider, $urlRouterProvider) {
        $urlRouterProvider.otherwise(function() {
            console.log("Route not found");
            window.location.replace("/");
        });
        $stateProvider
            .state("main", {
                url: "",
                views: {
                    '': {
                        templateUrl: 'static/templates/layout.html',
                        controller: "MainController",
                    },
                    "left1@main": {
                        templateUrl: "static/templates/center-menu.html",
                    },
                    "right@main": {
                        templateUrl: "static/templates/center-text.html",
                    },
                    "right2@main": {},
                    "topbar@main": {
                        templateUrl: 'static/templates/topbar.html'
                    },
                    "homePlaceholders@main": {
                        templateUrl: "static/templates/center.html",
                        controller: "HomeController"
                    }
                }
            })
            .state("main.personal", {
                url: "/personal/",
                views: {
                    "left1@main": {
                        templateUrl: "static/templates/personal-menu.html",
                    },
                    "left2@main": {
                        templateUrl: "static/templates/datepicker.html",
                        controller: "PersonalContoller"
                    },
                    "left3@main": {},
                    "full@main": {
                        templateUrl: 'static/templates/personal.html',
                        controller: "PersonalContoller"
                    },
                    "full2@main": {},
                    "right@main": {},
                    "right2@main": {},
                    "homePlaceholders@main": {}
                }
            })
            .state("main.personal.images", {
                url: "images",
                views: {
                    "right@main": {
                        templateUrl: 'static/templates/images.html',
                        controller: "PersonalImagesControler"
                    },
                    "right2@main": {},
                    "left1@main": {
                        templateUrl: "static/templates/personal-menu.html",
                    },
                    "left2@main": {
                        templateUrl: "static/templates/datepicker.html",
                        controller: "PersonalContoller"
                    },
                    "left3@main": {},
                    "full@main": {},
                    "full2@main": {},
                    "homePlaceholders@main": {}
                }
            })
            .state("main.personal.videos", {
                url: "videos",
                views: {
                    "right@main": {
                        templateUrl: 'static/templates/videos.html',
                        controller: "PersonalVideoControler"
                    },
                    "right2@main": {},
                    "left1@main": {
                        templateUrl: "static/templates/personal-menu.html",
                    },
                    "left2@main": {
                        templateUrl: "static/templates/datepicker.html",
                        controller: "PersonalContoller"
                    },
                    "left3@main": {},
                    "full@main": {},
                    "full2@main": {},
                    "homePlaceholders@main": {}
                }
            })
            .state("main.personal.gps", {
                url: "gps",
                views: {
                    "right@main": {
                        templateUrl: 'static/templates/mapdiv.html',
                        controller: "PersonalGPSControler"
                    },
                    "right2@main": {},
                    "left1@main": {
                        templateUrl: "static/templates/personal-menu.html",
                    },
                    "left2@main": {
                        templateUrl: "static/templates/datepicker.html",
                        controller: "PersonalContoller"
                    },
                    "left3@main": {},
                    "full@main": {},
                    "full2@main": {},
                    "homePlaceholders@main": {}
                }
            })
            .state("main.personal.images_gps", {
                url: "image_gps",
                views: {
                    "right@main": {
                        templateUrl: 'static/templates/img_gps.html',
                        controller: "PersonalImageGPSControler"
                    },
                    "right2@main": {},
                    "left1@main": {
                        templateUrl: "static/templates/personal-menu.html",
                    },
                    "left2@main": {
                        templateUrl: "static/templates/datepicker.html",
                        controller: "PersonalContoller"
                    },
                    "left3@main": {},
                    "full@main": {},
                    "full2@main": {},
                    "homePlaceholders@main": {}
                }
            })
            .state("main.personal.recommendation", {
                url: "recommendation",
                views: {
                    "right@main": {
                        templateUrl: 'static/templates/recommendation.html',
                        controller: "PersonalGPSControler"
                    },
                    "left1@main": {
                        templateUrl: "static/templates/personal-menu.html",
                    },
                    "left2@main": {
                        templateUrl: "static/templates/datepicker.html",
                        controller: "PersonalContoller"
                    },
                    "left3@main": {},
                    "full@main": {},
                    "full2@main": {}
                }
            })
            .state("main.personal.keymoments", {
                url: "keymoments",
                views: {
                    "right@main": {
                        templateUrl: 'static/templates/keymoments.html',
                        controller: "PersonalKeyMomentsController"
                    },
                    "left1@main": {
                        templateUrl: "static/templates/personal-menu.html",
                    },
                    "left2@main": {
                        templateUrl: "static/templates/datepicker.html",
                        controller: "PersonalContoller"
                    },
                    "left3@main": {},
                    "full@main": {},
                    "full2@main": {}
                }
            })
            .state("main.upload", {
                url: "/upload/",
                views: {
                    "full@main": {
                        templateUrl: 'static/templates/upload.html',
                        controller: "UploadController"
                    },
                    "right@main": {},
                    "right2@main": {},
                    "left1@main": {},
                    "left2@main": {},
                    "left3@main": {},
                    "full2@main": {},
                    "homePlaceholders@main": {}
                }
            })
            .state("main.fblogin", {
                url: "/fblogin/",
                views: {
                    "full@main": {
                        templateUrl: "static/templates/fblogintemplate.html",
                        controller: "FacebookLoginController"
                    }
                }
            })
            .state("main.stats", {
                url: "/stats/",
                views: {
                    "full2@main": {
                        templateUrl: 'static/templates/stats.html'
                    },
                    "right@main": {
                        templateUrl: "static/templates/imagecountbarchart.html"
                    },
                    "right2@main": {
                        templateUrl: "static/templates/gpscountbarchart.html"
                    },
                    "left1@main": {
                        templateUrl: "static/templates/datepicker.html",
                        controller: "PersonalContoller"
                    },
                    "left2@main": {
                        templateUrl: 'static/templates/statspanels.html',
                        controller: "StatsController"
                    },
                    "left3@main": {},
                    "full@main": {},
                    "homePlaceholders@main": {}
                }
            })
            .state("main.pstats", {
                url: "/pstats/",
                views: {
                    "full2@main": {
                        templateUrl: 'static/templates/stats.html'
                    },
                    "right@main": {
                        templateUrl: "static/templates/imagecountbarchart.html"
                    },
                    "right2@main": {
                        templateUrl: "static/templates/gpscountbarchart.html"
                    },
                    "left1@main": {
                        templateUrl: "static/templates/datepicker.html",
                        controller: "PersonalContoller"
                    },
                    "left2@main": {
                        templateUrl: "static/templates/personalstatspanels.html",
                        controller: "StatsController"
                    },
                    "left3@main": {},
                    "full@main": {},
                    "homePlaceholders@main": {}
                }
            })
    }
]);

app.controller("FacebookLoginController", ["$location", "$cookies", "$state",
    function($location, $cookies, $state) {
        var token = $location.search().token;

        if (!token)
            $location.url("/");
        $cookies.token = token;
        $location.url("/");
    }
]);

app.controller('UploadController', ['$scope', '$upload', "$timeout",
    function($scope, $upload, $timeout) {
        $scope.validImageTypes = ["image/jpeg", "image/png"];

        $scope.upload_progress = [];
        $scope.selectedFiles = [];
        $scope.upload_type = "";

        $scope.toBeUploaded = 0;
        $scope.doneUploading = 0;

        function dataURItoBlob(dataURI) {
            // convert base64/URLEncoded data component to raw binary data held in a string
            var byteString;
            if (dataURI.split(',')[0].indexOf('base64') >= 0)
                byteString = atob(dataURI.split(',')[1]);
            else
                byteString = unescape(dataURI.split(',')[1]);
            // separate out the mime component
            var mimeString = dataURI.split(',')[0].split(':')[1].split(';')[0];
            // write the bytes of the string to a typed array
            var ia = new Uint8Array(byteString.length);
            for (var i = 0; i < byteString.length; i++) {
                ia[i] = byteString.charCodeAt(i);
            }
            return new Blob([ia], {
                type: mimeString
            });
        }


        function loaded(evt) {
            // Obtain the read file data 
            var fileString = evt.target.result;
            $scope.selectedFiles.push({
                dataStr: fileString,
                skip: false
            });
            $scope.$apply();
        }

        $scope.getSelected = function() {
            return $scope.selectedFiles;
        }

        $scope.upload_file = function($files, type) {
            if (type === "image") {
                for (var i = 0; i < $files.length; i++) {
                    var reader = new FileReader();
                    reader.onload = loaded;
                    reader.readAsDataURL($files[i]);
                }
                $scope.upload_type = type;
            } else {
                $scope.toBeUploaded = $files.length;
                $scope.show_progress = true;
                for (var i = 0; i < $files.length; i++) {
                    $scope.start_upload(i, $files[i], type);
                }
            }
        }

        $scope.send_files = function() {
            var invalid_files = [];
            for (var i = 0; i < $scope.selectedFiles.length; i++) {
                if ($scope.selectedFiles[i].skip === "true" || $scope.selectedFiles[i].skip === true)
                    continue;
                $scope.toBeUploaded++;
            }
            for (var i = 0; i < $scope.selectedFiles.length; i++) {
                if ($scope.selectedFiles[i].skip === "true" || $scope.selectedFiles[i].skip === true)
                    continue;
                var file = $scope.selectedFiles[i].dataStr;
                $scope.start_upload(i, dataURItoBlob(file), $scope.upload_type);
            }
        }

        $scope.start_upload = function(i, file, type) {
            $scope.upload_progress.push({
                name: file.name,
                progress: 0
            });
            var upload = $upload.upload({
                url: "/api/upload/" + type,
                file: file,
            }).progress(function(evt) {
                $scope.upload_progress[i].progress = parseInt(100.0 * evt.loaded / evt.total);
            }).success(function(data, status, headers, config) {

                $scope.doneUploading++;
                $timeout(function() {
                    $scope.upload_progress[i].progress = 0;
                }, 1000);

                if ($scope.doneUploading === $scope.toBeUploaded) {
                    $scope.doneUploading = 0;
                    $scope.toBeUploaded = 0;
                    $scope.show_progress = false;
                    $scope.done = true;
                    $scope.newAlert("success", "Done uploading");
                    if ($scope.fileReference)
                        $scope.fileReference.splice(0, $scope.fileReference.length);
                    $scope.selectedFiles.splice(0, $scope.selectedFiles.length);
                }
            }).error(function(error) {
                console.log(error)
                $scope.newAlert("error", error);
            });
        }
    }
]);

app.controller("PersonalImagesControler", ["$scope", "Page",
    function($scope, Page) {
        $scope.classColor = "panel bluePanel";
        Page.reset();

        $scope.nextPage = function() {
            Page.nextPage();
            $scope.getImages();
        }

        $scope.previousPage = function() {
            Page.previousPage();
            $scope.getImages();
        }
        $scope.hasNext = Page.hasNext;
        $scope.hasPrevious = Page.hasPrevious;
        $scope.getCurrentPage = Page.getCurrentPage;

        $scope.attachAdditionalVariables = function(images) {
            images.forEach(function(image) {
                image.visible = false;
            });
            if (images.length > 0)
                images[0].visible = true;
        }

        $scope.setAllImages(true);
        $scope.setKeyMoments(false);
        console.log("key moments is in images", $scope.keyMoments);
        $scope.setPersonal(true);
        Page.setType("image");
        $scope.getImages();

    }
]);

app.controller("StatsController", ["$scope", "$location", "Page",
    function($scope, $location, Page) {
        $scope.classColor = "panel bluePanel";

        if (!String.prototype.startsWith) {
            String.prototype.startsWith = function(str) {
                return !this.indexOf(str);
            }
        }

        var url = $location.url();
        if (url.startsWith("/stats/"))
            Page.setType("stats");
        else if (url.startsWith("/pstats/"))
            Page.setType("pstats");

        $scope.pageType = Page.getType();
        $scope.getStats();
    }
]);

app.controller("PStatsController", ["$scope", "$location", "Page",
    function($scope, $location, Page) {
        $scope.classColor = "panel bluePanel";

        Page.setType("pstats");
        $scope.getStats();
    }
]);

app.controller("PersonalGPSControler", ["$scope", "Page",
    function($scope, Page) {
        $scope.classColor = "panel bluePanel";
        Page.reset();
        $scope.setPersonal(true);
        Page.setType("gps");
    }
]);

app.controller("PersonalImageGPSControler", ["$scope", "Page",
    function($scope, Page) {
        $scope.classColor = "panel bluePanel";
        Page.reset();
        $scope.setPersonal(true);
        Page.setType("image_gps");
        $scope.getImages();
    }
]);

app.controller("PersonalKeyMomentsController", ["$scope", "Page",
    function($scope, Page) {
        Page.reset();
        $scope.setKeyMoments(true);
        $scope.setAllImages(false);

        $scope.nextPage = function() {
            Page.nextPage();
            $scope.getImages();
        }

        $scope.previousPage = function() {
            Page.previousPage();
            $scope.getImages();
        }
        $scope.hasNext = Page.hasNext;
        $scope.hasPrevious = Page.hasPrevious;
        $scope.getCurrentPage = Page.getCurrentPage;

        $scope.attachAdditionalVariables = function(images) {
            images.forEach(function(image) {
                image.visible = false;
            });
            if (images.length > 0)
                images[0].visible = true;
        }
        $scope.setPersonal(true);
        Page.setType("image");
        $scope.getImages();
    }
]);

app.controller("PersonalVideoControler", ["$scope", "Page",
    function($scope, Page) {
        $scope.classColor = "panel bluePanel";
        Page.reset();

        $scope.nextPage = function() {
            Page.nextPage();
            $scope.getVideos();
        }

        $scope.hasNext = Page.hasNext;
        $scope.hasPrevious = Page.hasPrevious;
        $scope.getCurrentPage = Page.getCurrentPage;

        $scope.previousPage = function() {
            Page.previousPage();
            $scope.getVideos();
        }

        $scope.setPersonal(true);
        Page.setType("video");
        $scope.getVideos();
    }
]);

app.controller("PersonalContoller", ['$scope', "Page",
    function($scope, Page) {
        $scope.setPersonal(true);
        $scope.classColor = "panel darkbluePanel";
    }
]);

app.controller('HomeController', ['$scope', '$http',
    function($scope, $upload) {
        $scope.heatmap = true;

        $scope.classColor = "panel bluePanel";
    }
]);

app.filter('capitalize', function() {
    return function(input, all) {
        return (!!input) ? input.replace(/([^\W_]+[^\s-]*) */g,
            function(txt) {
                return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase();
            }) : '';
    }
});

// url is a string (http://localhost/) param has to be an object
function attachParam(url, param) {
    var keys = Object.keys(param);
    if (keys.length > 0)
        url = url + "?";
    keys.forEach(function(key) {
        url = url + key + "=" + param[key] + "&";
    });
    return url;
}

function extractGpsFromImages(images) {
    var parsedData = [];
    images.forEach(function(image) {
        if (image.lat && image.lon) {
            parsedData.push([image.lat, image.lon]);
        }
    });
    return parsedData;
}

function parseImages(images) {
    if (!Array.isArray(images))
        return images;
    var currentDate = "";
    var parsedData = {};
    parsedData.raw = angular.copy(images);
    var keys = [];
    images.forEach(function(image) {
        if (!currentDate || currentDate !== image.date.split("T")[0]) {
            currentDate = image.date.split("T")[0];
            keys.push(currentDate);
            parsedData[currentDate] = [image];
        } else {
            parsedData[currentDate].push(image);
        }
    });
    parsedData.keys = keys;
    return parsedData;
}

$(function() {
    $('[data-toggle="tooltip"]').tooltip()
});