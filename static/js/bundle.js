/******/ (function(modules) { // webpackBootstrap
/******/ 	// The module cache
/******/ 	var installedModules = {};

/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {

/******/ 		// Check if module is in cache
/******/ 		if(installedModules[moduleId])
/******/ 			return installedModules[moduleId].exports;

/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = installedModules[moduleId] = {
/******/ 			exports: {},
/******/ 			id: moduleId,
/******/ 			loaded: false
/******/ 		};

/******/ 		// Execute the module function
/******/ 		modules[moduleId].call(module.exports, module, module.exports, __webpack_require__);

/******/ 		// Flag the module as loaded
/******/ 		module.loaded = true;

/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}


/******/ 	// expose the modules object (__webpack_modules__)
/******/ 	__webpack_require__.m = modules;

/******/ 	// expose the module cache
/******/ 	__webpack_require__.c = installedModules;

/******/ 	// __webpack_public_path__
/******/ 	__webpack_require__.p = "";

/******/ 	// Load entry module and return exports
/******/ 	return __webpack_require__(0);
/******/ })
/************************************************************************/
/******/ ([
/* 0 */
/***/ function(module, exports, __webpack_require__) {

	var BarcodeReader = __webpack_require__(1)

	BarcodeReader.Init();

	BarcodeReader.DecodeSingleBarcode();

	BarcodeReader.SetImageCallback(function(result) {
	    if(result.length == 1) {
	        var ean = result[0].Value;
	        $("#barcode").val(ean); // Only take 1st bar code
	        App.isLoading = false; // hack to prevent loading disappearing
	        App.submit(ean);
	    } else {
	        App.stopLoading();
	        $("#recycle_form").addClass("has-error");
	        if(result.length == 0) {
	            $("#help").text("No bar code found");
	        } else {
	            $("#help").text("Multiple bar codes found");
	        }
	    }
	});

	BarcodeReader.SetImageErrorCallback(function() {
	    App.stopLoading();
	    $("#recycle_form").addClass("has-error");
	    $("#help").text("cannot read image");
	});

	var App = {
	    init: function(job) {
	        this.attachListeners();
	        this.product = null;
	        this.job = job;
	        this.isLoading = false;
	    },

	    attachListeners: function() {
	        var self = this;

	        $("#barcode_file_input").on("change", function(e){
	            $("#help").empty();
	            $("#recycle_form").removeClass("has-error");
	            $("#barcode").val("");
	            var input = $("#barcode_file_input");
	            if (input[0].files && input[0].files.length) {
	                var tmpImgURL = URL.createObjectURL(input[0].files[0]);
	                self.startLoading();
	                self.job.DecodeImage(tmpImgURL);
	            }
	        });


	        $("#recycle_form").submit(function (evt) {
	            evt.preventDefault();

	            var barcode = $("#barcode");
	            var barcodeVal = barcode.val();
	            return self.submit(barcodeVal);
	        });
	    },
	    detachListeners: function() {
	        $("#recycle_form").off("submit");
	    },

	    reset: function() {
	        $("#no_data").hide();
	        $("#product").hide();
	        $(".name").empty();
	        $(".ean").empty();
	        $("#image").src = "";
	        $("#throwaway").hide();
	        $("#bins").empty()
	        $("#source").empty();
	    },

	    startLoading: function() {
	        this.isLoading = true;
	        $("#loading").fadeTo("fast", 1.);
	    },

	    stopLoading: function() {
	        this.isLoading = false;
	        $("#loading").fadeTo("fast", 0.);
	    },

	    submit: function(ean) {
	        var self = this;

	        if (self.isLoading) {
	            return false;
	        }

	        self.reset();

	        var recycle = $("#recycle_form");
	        if (ean.length == 0) {
	            recycle.addClass("has-error");
	        } else {
	            self.product = null;
	            self.startLoading();
	            recycle.removeClass("has-error");
	            $("#help").empty();
	            var url = "/throwaway/" + ean;

	            $.get(url, function(data) {
	                data = $.parseJSON(data)
	                var throwAway = data.throwAway;
	                self.product = data.product;
	                $(".name").text(self.product.name);
	                $(".ean").text(self.product.ean);
	                $("#image").attr("src", self.product.image_url);
	                var source = $("#source");
	                if (self.product.website_url == "") {
	                    source.text("Source: " + self.product.website_name)
	                } else {
	                    source.html("Source: " + '<a href="' + self.product.website_url + '" rel="noopener" target="_blank">' + self.product.website_name + '</a>');
	                }
	                $("#product").show();

	                if (jQuery.isEmptyObject(throwAway)) {
	                    $("#no_data").show();
	                    self.stopLoading();
	                } else {
	                    var binCount = 0;
	                    var bins = {};
	                    for (var k in throwAway) {
	                        binCount++;
	                        var bin = throwAway[k];
	                        if (bin in bins) {
	                            bins[bin].push(k);
	                        } else {
	                            bins[bin] = [k];
	                        }
	                    }

	                    var binDiv = $("#bins");
	                    binDiv.empty();
	                    var binCls = "col-sm-" + parseInt(12 / binCount);

	                    var binId = 0;
	                    for (var k in bins) {
	                        binDiv.append('<div class="text-center"><h3>' + k + '</h3><div id="bin-' + parseInt(binId) + '"></div></div>');
	                        var bin = $("#bin-" + parseInt(binId));
	                        var materials = bins[k];
	                        for (var i = 0; i < materials.length; i++) {
	                            bin.append(materials[i] + "<br>");
	                        }
	                        binId++;
	                    }
	                    $("#throwaway").show();
	                    self.stopLoading();
	                }
	            }).fail(function(xhr) {
	                recycle.addClass("has-error");
	                $("#help").html(xhr.responseText.replace(/\n/g, "<br>"));
	                self.stopLoading();
	            });
	        }
	        return false;
	    }
	};

	App.init(BarcodeReader);

	var Suggester = {
	    init: function(app) {
	        this.attachListeners();
	        this.app = app;
	    },
	    attachListeners: function() {
	        var self = this;
	        var suggest_form = $("#suggest_form");

	        $("#suggest_modal").on("hidden.bs.modal", function () {
	            suggest_form.removeClass("has-error");
	        });

	        suggest_form.submit(function (evt) {
	            evt.preventDefault();

	            var selected = [];
	            $('#materials_list input:checked').each(function() {
	                var name = $(this).attr("id");
	                var id = name.replace("materialId-", "");
	                var text = $("label[for=" + name + "]");
	                selected.push({name: text.text(), id: parseInt(id)});
	            });

	            if (selected.length == 0) {
	                suggest_form.addClass("has-error");
	            } else {
	                $('#suggest_modal').modal('toggle');
	                 $.post("/package/add", {materials: JSON.stringify(selected), ean: self.app.product.ean}, function() {
	                    self.app.submit(self.app.product.ean);
	                });
	            }

	            return false;
	        });

	        $(".suggest_button").click(function() {
	            var url = "/materials/";
	            var lst = $("#materials_list");
	            lst.html("");

	            $.get(url, function(data) {
	                data = $.parseJSON(data)

	                if (data.length == 0) {
	                    $("#suggest_help").text("Pas d'emballage disponible");
	                } else {
	                    for (var i = 0; i < data.length; i++) {
	                        var v = data[i]
	                        self.addMaterialItem(lst, v.id, v.name);
	                    }

	                    if (self.app.product != null) {
	                        for (var i = 0; i < self.app.product.materials.length; i++) {
	                            var id = self.app.product.materials[i].id;
	                            $("#materialId-" + id.toString()).prop("checked", true);
	                        }
	                    }
	                }
	            }).fail(function(xhr) {
	                $("#suggest_help").html(xhr.responseText.replace(/\n/g, "<br>"));
	            });
	        });
	    },

	    addMaterialItem: function(lst, id, name) {
	        lst.append('<div class="checkbox col-sm-4 suggest-checkbox"><label for="materialId-' + id + '"><input type="checkbox" value="" id="materialId-' + id + '">' + name + '</label></div>');
	    },

	    detachListeners: function() {
	        $("#suggest_modal").off("hidden.bs.modal");
	        $("#suggest_form").off("submit");
	        $(".suggest_button").off("click");
	    },

	};

	Suggester.init(App);


	var BlackLister = {
	    init: function(app) {
	        this.attachListeners();
	        this.app = app;
	    },

	    attachListeners: function() {
	        var self = this;

	        var blacklist = $("#blacklist_form");
	        $("#blacklist_modal").on("hidden.bs.modal", function () {
	            blacklist.removeClass("has-error");
	        });

	        blacklist.submit(function (evt) {
	            evt.preventDefault();
	            var name = $("#blacklist_name");
	            var nameVal = name.val();

	            if (nameVal.length == 0) {
	                blacklist.addClass("has-error");
	            } else {
	                var product = self.app.product;
	                $('#blacklist_modal').modal('toggle');
	                $.post("/blacklist/add", {name: nameVal, url: product.url, ean: product.ean, website: product.website_name}, function() {
	                    self.app.submit(self.app.product.ean);
	                });
	            }

	            return false;
	        });
	    },

	    detachListeners: function() {
	        $("#blacklist_modal").off("hidden.bs.modal");
	        $("#blacklist_form").off("submit");
	    },

	};

	BlackLister.init(App);


/***/ },
/* 1 */
/***/ function(module, exports, __webpack_require__) {

	/**
	 * CallBacks:
	 * __________________________________________________________________________________
	 * All the callback function should have one parameter:
	 * function(result){};
	 * And the result parameter will contain an array of objects that look like BarcodeReader.
	 * result = [{Format: the barcode type, Value: the value of the barcode}];
	 * __________________________________________________________________________________
	 *
	 * You can use either the set functions or just access the properties directly to set callback or
	 * other properties. Just always remember to call Init() before starting to decode something never mess
	 * around with the SupportedFormats property.
	 *
	 */

	var EXIF = __webpack_require__(2);
	var decoderWorkerBlobString = __webpack_require__(3);

	var BarcodeReader = {
	  Config: {
	    // Set to false if the decoder should look for one barcode and then stop. Increases performance.
	    Multiple: true,

	    // The formats that the decoder will look for.
	    DecodeFormats: ["Code128", "Code93", "Code39", "EAN-13", "2Of5", "Inter2Of5", "Codabar"],

	    // ForceUnique just must makes sure that the callback function isn't repeatedly called
	    // with the same barcode. Especially in the case of a video stream.
	    ForceUnique: true,

	    // Set to true if information about the localization should be recieved from the worker.
	    LocalizationFeedback: false,

	    // Set to true if checking orientation of the image should be skipped.
	    // Checking orientation takes a bit of time for larger images, so if
	    // you are sure that the image orientation is 1 you should skip it.
	    SkipOrientation: false
	  },
	  SupportedFormats: ["Code128", "Code93", "Code39", "EAN-13", "2Of5", "Inter2Of5", "Codabar"], // Don't touch.
	  ScanCanvas: null, // Don't touch the canvas either.
	  ScanContext: null,
	  SquashCanvas: document.createElement("canvas"),
	  ImageCallback: null, // Callback for the decoding of an image.
	  StreamCallback: null, // Callback for the decoding of a video.
	  LocalizationCallback: null, // Callback for localization.
	  ImageErrorCallback: null, // Callback for error on image loading.
	  Stream: null, // The actual video.
	  DecodeStreamActive: false, // Will be set to false when StopStreamDecode() is called.
	  Decoded: [], // Used to enfore the ForceUnique property.
	  DecoderWorker: new Worker( URL.createObjectURL(new Blob([decoderWorkerBlobString], {type: "application/javascript"}) ) ),
	  OrientationCallback: null,
	  // Always call the Init().
	  Init: function() {
	    BarcodeReader.ScanCanvas = BarcodeReader.FixCanvas(document.createElement("canvas"));
	    BarcodeReader.ScanCanvas.width = 640;
	    BarcodeReader.ScanCanvas.height = 480;
	    BarcodeReader.ScanContext = BarcodeReader.ScanCanvas.getContext("2d");
	  },

	  // Value should be true or false.
	  SetRotationSkip: function(value) {
	    BarcodeReader.Config.SkipOrientation = value;
	  },
	  // Sets the callback function for the image decoding.
	  SetImageCallback: function(callBack) {
	    BarcodeReader.ImageCallback = callBack;
	  },

	  // Sets the callback function for the video decoding.
	  SetStreamCallback: function(callBack) {
	    BarcodeReader.StreamCallback = callBack;
	  },

	  // Sets callback for localization, the callback function should take one argument.
	  // This will be an array with objects with format.
	  // {x, y, width, height}
	  // This represents a localization rectangle.
	  // The rectangle comes from a 320, 240 area i.e the search canvas.
	  SetLocalizationCallback: function(callBack) {
	    BarcodeReader.LocalizationCallback = callBack;
	    BarcodeReader.Config.LocalizationFeedback = true;
	  },

	  // Sets the callback function when loading a wrong image.
	  SetImageErrorCallback: function(callBack) {
	    BarcodeReader.ImageErrorCallback = callBack;
	  },

	  // Set to true if LocalizationCallback is set and you would like to
	  // receive the feedback or false if
	  SwitchLocalizationFeedback: function(bool) {
	    BarcodeReader.Config.LocalizationFeedback = bool;
	  },

	  // Switches for changing the Multiple property.
	  DecodeSingleBarcode: function() {
	    BarcodeReader.Config.Multiple = false;
	  },
	  DecodeMultiple: function() {
	    BarcodeReader.Config.Multiple = true;
	  },

	  // Sets the formats to decode, formats should be an array of a subset of the supported formats.
	  SetDecodeFormats: function(formats) {
	    BarcodeReader.Config.DecodeFormats = [];
	    for (var i = 0; i < formats.length; i++) {
	      if (BarcodeReader.SupportedFormats.indexOf(formats[i]) !== -1) {
	        BarcodeReader.Config.DecodeFormats.push(formats[i]);
	      }
	    }
	    if (BarcodeReader.Config.DecodeFormats.length === 0) {
	      BarcodeReader.Config.DecodeFormats = BarcodeReader.SupportedFormats.slice();
	    }
	  },

	  // Removes a list of formats from the formats to decode.
	  SkipFormats: function(formats) {
	    for (var i = 0; i < formats.length; i++) {
	      var index = BarcodeReader.Config.DecodeFormats.indexOf(formats[i]);
	      if (index >= 0) {
	        BarcodeReader.Config.DecodeFormats.splice(index, 1);
	      }
	    }
	  },

	  // Adds a list of formats to the formats to decode.
	  AddFormats: function(formats) {
	    for (var i = 0; i < formats.length; i++) {
	      if (BarcodeReader.SupportedFormats.indexOf(formats[i]) !== -1) {
	        if (BarcodeReader.Config.DecodeFormats.indexOf(formats[i]) === -1) {
	          BarcodeReader.Config.DecodeFormats.push(formats[i]);
	        }
	      }
	    }
	  },

	  // The callback function for image decoding used internally by BarcodeReader.
	  BarcodeReaderImageCallback: function(e) {
	    if (e.data.success === "localization") {
	      if (BarcodeReader.Config.LocalizationFeedback) {
	        BarcodeReader.LocalizationCallback(e.data.result);
	      }
	      return;
	    }
	    if (e.data.success === "orientationData") {
	      BarcodeReader.OrientationCallback(e.data.result);
	      return;
	    }
	    var filteredData = [];
	    for (var i = 0; i < e.data.result.length; i++) {
	      if (BarcodeReader.Decoded.indexOf(e.data.result[i].Value) === -1 || BarcodeReader.Config.ForceUnique === false) {
	        filteredData.push(e.data.result[i]);
	        if (BarcodeReader.Config.ForceUnique) BarcodeReader.Decoded.push(e.data.result[i].Value);
	      }
	    }
	    BarcodeReader.ImageCallback(filteredData);
	    BarcodeReader.Decoded = [];
	  },

	  // The callback function for stream decoding used internally by BarcodeReader.
	  BarcodeReaderStreamCallback: function(e) {
	    if (e.data.success === "localization") {
	      if (BarcodeReader.Config.LocalizationFeedback) {
	        BarcodeReader.LocalizationCallback(e.data.result);
	      }
	      return;
	    }
	    if (e.data.success && BarcodeReader.DecodeStreamActive) {
	      var filteredData = [];
	      for (var i = 0; i < e.data.result.length; i++) {
	        if (BarcodeReader.Decoded.indexOf(e.data.result[i].Value) === -1 || BarcodeReader.ForceUnique === false) {
	          filteredData.push(e.data.result[i]);
	          if (BarcodeReader.ForceUnique) BarcodeReader.Decoded.push(e.data.result[i].Value);
	        }
	      }
	      if (filteredData.length > 0) {
	        BarcodeReader.StreamCallback(filteredData);
	      }
	    }
	    if (BarcodeReader.DecodeStreamActive) {
	      BarcodeReader.ScanContext.drawImage(BarcodeReader.Stream, 0, 0, BarcodeReader.ScanCanvas.width, BarcodeReader.ScanCanvas.height);
	      BarcodeReader.DecoderWorker.postMessage({
	        scan: BarcodeReader.ScanContext.getImageData(0, 0, BarcodeReader.ScanCanvas.width, BarcodeReader.ScanCanvas.height).data,
	        scanWidth: BarcodeReader.ScanCanvas.width,
	        scanHeight: BarcodeReader.ScanCanvas.height,
	        multiple: BarcodeReader.Config.Multiple,
	        decodeFormats: BarcodeReader.Config.DecodeFormats,
	        cmd: "normal",
	        rotation: 1
	      });

	    }
	    if (!BarcodeReader.DecodeStreamActive) {
	      BarcodeReader.Decoded = [];
	    }
	  },

	  // The image decoding function, image is a data source for an image or an image element.
	  DecodeImage: function(image) {
		var img = new Image();
		img.onerror = BarcodeReader.ImageErrorCallback;

	    if (image instanceof Image || image instanceof HTMLImageElement) {
	      image.exifdata = false;
	      if (image.complete) {
	        if (BarcodeReader.Config.SkipOrientation) {
	          BarcodeReader.BarcodeReaderDecodeImage(image, 1, "");
	        } else {
	          EXIF.getData(image, function(exifImage) {
	            var orientation = EXIF.getTag(exifImage, "Orientation");
	            var sceneType = EXIF.getTag(exifImage, "SceneCaptureType");
	            if (typeof orientation !== 'number') orientation = 1;
	            BarcodeReader.BarcodeReaderDecodeImage(exifImage, orientation, sceneType);
	          });
	        }
	      } else {
	        img.onload = function() {
	          if (BarcodeReader.Config.SkipOrientation) {
	            BarcodeReader.BarcodeReaderDecodeImage(img, 1, "");
	          } else {
	            EXIF.getData(this, function(exifImage) {
	              var orientation = EXIF.getTag(exifImage, "Orientation");
	              var sceneType = EXIF.getTag(exifImage, "SceneCaptureType");
	              if (typeof orientation !== 'number') orientation = 1;
	              BarcodeReader.BarcodeReaderDecodeImage(exifImage, orientation, sceneType);
	            });
	          }
	        };
	        img.src = image.src;
	      }
	    } else {
	      img.onload = function() {
	        if (BarcodeReader.Config.SkipOrientation) {
	          BarcodeReader.BarcodeReaderDecodeImage(img, 1, "");
	        } else {
	          EXIF.getData(this, function(exifImage) {
	            var orientation = EXIF.getTag(exifImage, "Orientation");
	            var sceneType = EXIF.getTag(exifImage, "SceneCaptureType");
	            if (typeof orientation !== 'number') orientation = 1;
	            BarcodeReader.BarcodeReaderDecodeImage(exifImage, orientation, sceneType);
	          });
	        }
	      };
	      img.src = image;
	    }
	  },

	  // Starts the decoding of a stream, the stream is a video not a blob i.e it's an element.
	  DecodeStream: function(stream) {
	    BarcodeReader.Stream = stream;
	    BarcodeReader.DecodeStreamActive = true;
	    BarcodeReader.DecoderWorker.onmessage = BarcodeReader.BarcodeReaderStreamCallback;
	    BarcodeReader.ScanContext.drawImage(stream, 0, 0, BarcodeReader.ScanCanvas.width, BarcodeReader.ScanCanvas.height);
	    BarcodeReader.DecoderWorker.postMessage({
	      scan: BarcodeReader.ScanContext.getImageData(0, 0, BarcodeReader.ScanCanvas.width, BarcodeReader.ScanCanvas.height).data,
	      scanWidth: BarcodeReader.ScanCanvas.width,
	      scanHeight: BarcodeReader.ScanCanvas.height,
	      multiple: BarcodeReader.Config.Multiple,
	      decodeFormats: BarcodeReader.Config.DecodeFormats,
	      cmd: "normal",
	      rotation: 1
	    });
	  },

	  // Stops the decoding of a stream.
	  StopStreamDecode: function() {
	    BarcodeReader.DecodeStreamActive = false;
	    BarcodeReader.Decoded = [];
	  },

	  BarcodeReaderDecodeImage: function(image, orientation, sceneCaptureType) {
	    if (orientation === 8 || orientation === 6) {
	      if (sceneCaptureType === "Landscape" && image.width > image.height) {
	        orientation = 1;
	        BarcodeReader.ScanCanvas.width = 640;
	        BarcodeReader.ScanCanvas.height = 480;
	      } else {
	        BarcodeReader.ScanCanvas.width = 480;
	        BarcodeReader.ScanCanvas.height = 640;
	      }
	    } else {
	      BarcodeReader.ScanCanvas.width = 640;
	      BarcodeReader.ScanCanvas.height = 480;
	    }
	    BarcodeReader.DecoderWorker.onmessage = BarcodeReader.BarcodeReaderImageCallback;
	    BarcodeReader.ScanContext.drawImage(image, 0, 0, BarcodeReader.ScanCanvas.width, BarcodeReader.ScanCanvas.height);
	    BarcodeReader.Orientation = orientation;
	    BarcodeReader.DecoderWorker.postMessage({
	      scan: BarcodeReader.ScanContext.getImageData(0, 0, BarcodeReader.ScanCanvas.width, BarcodeReader.ScanCanvas.height).data,
	      scanWidth: BarcodeReader.ScanCanvas.width,
	      scanHeight: BarcodeReader.ScanCanvas.height,
	      multiple: BarcodeReader.Config.Multiple,
	      decodeFormats: BarcodeReader.Config.DecodeFormats,
	      cmd: "normal",
	      rotation: orientation,
	      postOrientation: BarcodeReader.PostOrientation
	    });
	  },

	  DetectVerticalSquash: function(img) {
	    var ih = img.naturalHeight;
	    var canvas = BarcodeReader.SquashCanvas;
	    var alpha;
	    var data;
	    canvas.width = 1;
	    canvas.height = ih;
	    var ctx = canvas.getContext('2d');
	    ctx.drawImage(img, 0, 0);
	    try {
	      data = ctx.getImageData(0, 0, 1, ih).data;
	    } catch (err) {
	      console.log("Cannot check verticalSquash: CORS?");
	      return 1;
	    }
	    var sy = 0;
	    var ey = ih;
	    var py = ih;
	    while (py > sy) {
	      alpha = data[(py - 1) * 4 + 3];
	      if (alpha === 0) {
	        ey = py;
	      } else {
	        sy = py;
	      }
	      py = (ey + sy) >> 1;
	    }
	    var ratio = (py / ih);
	    return (ratio === 0) ? 1 : ratio;
	  },

	  FixCanvas: function(canvas) {
	    var ctx = canvas.getContext('2d');
	    var drawImage = ctx.drawImage;
	    ctx.drawImage = function(img, sx, sy, sw, sh, dx, dy, dw, dh) {
	      var vertSquashRatio = 1;
	      if (!!img && img.nodeName === 'IMG') {
	        vertSquashRatio = BarcodeReader.DetectVerticalSquash(img);
	        // sw || (sw = img.naturalWidth);
	        // sh || (sh = img.naturalHeight);
	      }
	      if (arguments.length === 9)
	        drawImage.call(ctx, img, sx, sy, sw, sh, dx, dy, dw, dh / vertSquashRatio);
	      else if (typeof sw !== 'undefined')
	        drawImage.call(ctx, img, sx, sy, sw, sh / vertSquashRatio);
	      else
	        drawImage.call(ctx, img, sx, sy);
	    };
	    return canvas;
	  }
	};


	if (true) {
	  if (typeof module !== 'undefined' && module.exports) {
	      exports = module.exports = BarcodeReader;
	  }
	  exports.BarcodeReader = BarcodeReader;
	} else {
	  root.BarcodeReader = BarcodeReader;
	}


/***/ },
/* 2 */
/***/ function(module, exports, __webpack_require__) {

	var __WEBPACK_AMD_DEFINE_ARRAY__, __WEBPACK_AMD_DEFINE_RESULT__;(function() {

	  var debug = false;

	  var root = this;

	  var EXIF = function(obj) {
	    if (obj instanceof EXIF) return obj;
	    if (!(this instanceof EXIF)) return new EXIF(obj);
	    this.EXIFwrapped = obj;
	  };

	  if (true) {
	    if (typeof module !== 'undefined' && module.exports) {
	      exports = module.exports = EXIF;
	    }
	    exports.EXIF = EXIF;
	  } else {
	    root.EXIF = EXIF;
	  }

	  var ExifTags = EXIF.Tags = {

	    // version tags
	    0x9000: "ExifVersion", // EXIF version
	    0xA000: "FlashpixVersion", // Flashpix format version

	    // colorspace tags
	    0xA001: "ColorSpace", // Color space information tag

	    // image configuration
	    0xA002: "PixelXDimension", // Valid width of meaningful image
	    0xA003: "PixelYDimension", // Valid height of meaningful image
	    0x9101: "ComponentsConfiguration", // Information about channels
	    0x9102: "CompressedBitsPerPixel", // Compressed bits per pixel

	    // user information
	    0x927C: "MakerNote", // Any desired information written by the manufacturer
	    0x9286: "UserComment", // Comments by user

	    // related file
	    0xA004: "RelatedSoundFile", // Name of related sound file

	    // date and time
	    0x9003: "DateTimeOriginal", // Date and time when the original image was generated
	    0x9004: "DateTimeDigitized", // Date and time when the image was stored digitally
	    0x9290: "SubsecTime", // Fractions of seconds for DateTime
	    0x9291: "SubsecTimeOriginal", // Fractions of seconds for DateTimeOriginal
	    0x9292: "SubsecTimeDigitized", // Fractions of seconds for DateTimeDigitized

	    // picture-taking conditions
	    0x829A: "ExposureTime", // Exposure time (in seconds)
	    0x829D: "FNumber", // F number
	    0x8822: "ExposureProgram", // Exposure program
	    0x8824: "SpectralSensitivity", // Spectral sensitivity
	    0x8827: "ISOSpeedRatings", // ISO speed rating
	    0x8828: "OECF", // Optoelectric conversion factor
	    0x9201: "ShutterSpeedValue", // Shutter speed
	    0x9202: "ApertureValue", // Lens aperture
	    0x9203: "BrightnessValue", // Value of brightness
	    0x9204: "ExposureBias", // Exposure bias
	    0x9205: "MaxApertureValue", // Smallest F number of lens
	    0x9206: "SubjectDistance", // Distance to subject in meters
	    0x9207: "MeteringMode", // Metering mode
	    0x9208: "LightSource", // Kind of light source
	    0x9209: "Flash", // Flash status
	    0x9214: "SubjectArea", // Location and area of main subject
	    0x920A: "FocalLength", // Focal length of the lens in mm
	    0xA20B: "FlashEnergy", // Strobe energy in BCPS
	    0xA20C: "SpatialFrequencyResponse", //
	    0xA20E: "FocalPlaneXResolution", // Number of pixels in width direction per FocalPlaneResolutionUnit
	    0xA20F: "FocalPlaneYResolution", // Number of pixels in height direction per FocalPlaneResolutionUnit
	    0xA210: "FocalPlaneResolutionUnit", // Unit for measuring FocalPlaneXResolution and FocalPlaneYResolution
	    0xA214: "SubjectLocation", // Location of subject in image
	    0xA215: "ExposureIndex", // Exposure index selected on camera
	    0xA217: "SensingMethod", // Image sensor type
	    0xA300: "FileSource", // Image source (3 == DSC)
	    0xA301: "SceneType", // Scene type (1 == directly photographed)
	    0xA302: "CFAPattern", // Color filter array geometric pattern
	    0xA401: "CustomRendered", // Special processing
	    0xA402: "ExposureMode", // Exposure mode
	    0xA403: "WhiteBalance", // 1 = auto white balance, 2 = manual
	    0xA404: "DigitalZoomRation", // Digital zoom ratio
	    0xA405: "FocalLengthIn35mmFilm", // Equivalent foacl length assuming 35mm film camera (in mm)
	    0xA406: "SceneCaptureType", // Type of scene
	    0xA407: "GainControl", // Degree of overall image gain adjustment
	    0xA408: "Contrast", // Direction of contrast processing applied by camera
	    0xA409: "Saturation", // Direction of saturation processing applied by camera
	    0xA40A: "Sharpness", // Direction of sharpness processing applied by camera
	    0xA40B: "DeviceSettingDescription", //
	    0xA40C: "SubjectDistanceRange", // Distance to subject

	    // other tags
	    0xA005: "InteroperabilityIFDPointer",
	    0xA420: "ImageUniqueID" // Identifier assigned uniquely to each image
	  };

	  var TiffTags = EXIF.TiffTags = {
	    0x0100: "ImageWidth",
	    0x0101: "ImageHeight",
	    0x8769: "ExifIFDPointer",
	    0x8825: "GPSInfoIFDPointer",
	    0xA005: "InteroperabilityIFDPointer",
	    0x0102: "BitsPerSample",
	    0x0103: "Compression",
	    0x0106: "PhotometricInterpretation",
	    0x0112: "Orientation",
	    0x0115: "SamplesPerPixel",
	    0x011C: "PlanarConfiguration",
	    0x0212: "YCbCrSubSampling",
	    0x0213: "YCbCrPositioning",
	    0x011A: "XResolution",
	    0x011B: "YResolution",
	    0x0128: "ResolutionUnit",
	    0x0111: "StripOffsets",
	    0x0116: "RowsPerStrip",
	    0x0117: "StripByteCounts",
	    0x0201: "JPEGInterchangeFormat",
	    0x0202: "JPEGInterchangeFormatLength",
	    0x012D: "TransferFunction",
	    0x013E: "WhitePoint",
	    0x013F: "PrimaryChromaticities",
	    0x0211: "YCbCrCoefficients",
	    0x0214: "ReferenceBlackWhite",
	    0x0132: "DateTime",
	    0x010E: "ImageDescription",
	    0x010F: "Make",
	    0x0110: "Model",
	    0x0131: "Software",
	    0x013B: "Artist",
	    0x8298: "Copyright"
	  };

	  var GPSTags = EXIF.GPSTags = {
	    0x0000: "GPSVersionID",
	    0x0001: "GPSLatitudeRef",
	    0x0002: "GPSLatitude",
	    0x0003: "GPSLongitudeRef",
	    0x0004: "GPSLongitude",
	    0x0005: "GPSAltitudeRef",
	    0x0006: "GPSAltitude",
	    0x0007: "GPSTimeStamp",
	    0x0008: "GPSSatellites",
	    0x0009: "GPSStatus",
	    0x000A: "GPSMeasureMode",
	    0x000B: "GPSDOP",
	    0x000C: "GPSSpeedRef",
	    0x000D: "GPSSpeed",
	    0x000E: "GPSTrackRef",
	    0x000F: "GPSTrack",
	    0x0010: "GPSImgDirectionRef",
	    0x0011: "GPSImgDirection",
	    0x0012: "GPSMapDatum",
	    0x0013: "GPSDestLatitudeRef",
	    0x0014: "GPSDestLatitude",
	    0x0015: "GPSDestLongitudeRef",
	    0x0016: "GPSDestLongitude",
	    0x0017: "GPSDestBearingRef",
	    0x0018: "GPSDestBearing",
	    0x0019: "GPSDestDistanceRef",
	    0x001A: "GPSDestDistance",
	    0x001B: "GPSProcessingMethod",
	    0x001C: "GPSAreaInformation",
	    0x001D: "GPSDateStamp",
	    0x001E: "GPSDifferential"
	  };

	  var StringValues = EXIF.StringValues = {
	    ExposureProgram: {
	      0: "Not defined",
	      1: "Manual",
	      2: "Normal program",
	      3: "Aperture priority",
	      4: "Shutter priority",
	      5: "Creative program",
	      6: "Action program",
	      7: "Portrait mode",
	      8: "Landscape mode"
	    },
	    MeteringMode: {
	      0: "Unknown",
	      1: "Average",
	      2: "CenterWeightedAverage",
	      3: "Spot",
	      4: "MultiSpot",
	      5: "Pattern",
	      6: "Partial",
	      255: "Other"
	    },
	    LightSource: {
	      0: "Unknown",
	      1: "Daylight",
	      2: "Fluorescent",
	      3: "Tungsten (incandescent light)",
	      4: "Flash",
	      9: "Fine weather",
	      10: "Cloudy weather",
	      11: "Shade",
	      12: "Daylight fluorescent (D 5700 - 7100K)",
	      13: "Day white fluorescent (N 4600 - 5400K)",
	      14: "Cool white fluorescent (W 3900 - 4500K)",
	      15: "White fluorescent (WW 3200 - 3700K)",
	      17: "Standard light A",
	      18: "Standard light B",
	      19: "Standard light C",
	      20: "D55",
	      21: "D65",
	      22: "D75",
	      23: "D50",
	      24: "ISO studio tungsten",
	      255: "Other"
	    },
	    Flash: {
	      0x0000: "Flash did not fire",
	      0x0001: "Flash fired",
	      0x0005: "Strobe return light not detected",
	      0x0007: "Strobe return light detected",
	      0x0009: "Flash fired, compulsory flash mode",
	      0x000D: "Flash fired, compulsory flash mode, return light not detected",
	      0x000F: "Flash fired, compulsory flash mode, return light detected",
	      0x0010: "Flash did not fire, compulsory flash mode",
	      0x0018: "Flash did not fire, auto mode",
	      0x0019: "Flash fired, auto mode",
	      0x001D: "Flash fired, auto mode, return light not detected",
	      0x001F: "Flash fired, auto mode, return light detected",
	      0x0020: "No flash function",
	      0x0041: "Flash fired, red-eye reduction mode",
	      0x0045: "Flash fired, red-eye reduction mode, return light not detected",
	      0x0047: "Flash fired, red-eye reduction mode, return light detected",
	      0x0049: "Flash fired, compulsory flash mode, red-eye reduction mode",
	      0x004D: "Flash fired, compulsory flash mode, red-eye reduction mode, return light not detected",
	      0x004F: "Flash fired, compulsory flash mode, red-eye reduction mode, return light detected",
	      0x0059: "Flash fired, auto mode, red-eye reduction mode",
	      0x005D: "Flash fired, auto mode, return light not detected, red-eye reduction mode",
	      0x005F: "Flash fired, auto mode, return light detected, red-eye reduction mode"
	    },
	    SensingMethod: {
	      1: "Not defined",
	      2: "One-chip color area sensor",
	      3: "Two-chip color area sensor",
	      4: "Three-chip color area sensor",
	      5: "Color sequential area sensor",
	      7: "Trilinear sensor",
	      8: "Color sequential linear sensor"
	    },
	    SceneCaptureType: {
	      0: "Standard",
	      1: "Landscape",
	      2: "Portrait",
	      3: "Night scene"
	    },
	    SceneType: {
	      1: "Directly photographed"
	    },
	    CustomRendered: {
	      0: "Normal process",
	      1: "Custom process"
	    },
	    WhiteBalance: {
	      0: "Auto white balance",
	      1: "Manual white balance"
	    },
	    GainControl: {
	      0: "None",
	      1: "Low gain up",
	      2: "High gain up",
	      3: "Low gain down",
	      4: "High gain down"
	    },
	    Contrast: {
	      0: "Normal",
	      1: "Soft",
	      2: "Hard"
	    },
	    Saturation: {
	      0: "Normal",
	      1: "Low saturation",
	      2: "High saturation"
	    },
	    Sharpness: {
	      0: "Normal",
	      1: "Soft",
	      2: "Hard"
	    },
	    SubjectDistanceRange: {
	      0: "Unknown",
	      1: "Macro",
	      2: "Close view",
	      3: "Distant view"
	    },
	    FileSource: {
	      3: "DSC"
	    },

	    Components: {
	      0: "",
	      1: "Y",
	      2: "Cb",
	      3: "Cr",
	      4: "R",
	      5: "G",
	      6: "B"
	    }
	  };

	  function addEvent(element, event, handler) {
	    if (element.addEventListener) {
	      element.addEventListener(event, handler, false);
	    } else if (element.attachEvent) {
	      element.attachEvent("on" + event, handler);
	    }
	  }

	  function imageHasData(img) {
	    return !!(img.exifdata);
	  }


	  function base64ToArrayBuffer(base64, contentType) {
	    contentType = contentType || base64.match(/^data\:([^\;]+)\;base64,/mi)[1] || ''; // e.g. 'data:image/jpeg;base64,...' => 'image/jpeg'
	    base64 = base64.replace(/^data\:([^\;]+)\;base64,/gmi, '');
	    var binary = atob(base64);
	    var len = binary.length;
	    var buffer = new ArrayBuffer(len);
	    var view = new Uint8Array(buffer);
	    for (var i = 0; i < len; i++) {
	      view[i] = binary.charCodeAt(i);
	    }
	    return buffer;
	  }

	  function objectURLToBlob(url, callback) {
	    var http = new XMLHttpRequest();
	    http.open("GET", url, true);
	    http.responseType = "blob";
	    http.onload = function(e) {
	      if (this.status == 200 || this.status === 0) {
	        callback(this.response);
	      }
	    };
	    http.send();
	  }

	  function getImageData(img, callback) {
	    var fileReader = new FileReader();
	    var handleBinaryFile = function handleBinaryFile(binFile) {
	      var data = findEXIFinJPEG(binFile);
	      var iptcdata = findIPTCinJPEG(binFile);
	      img.exifdata = data || {};
	      img.iptcdata = iptcdata || {};
	      if (callback) {
	        callback(img);
	      }
	    };

	    if (img.src) {
	      if (/^data\:/i.test(img.src)) { // Data URI
	        var arrayBuffer = base64ToArrayBuffer(img.src);
	        handleBinaryFile(arrayBuffer);

	      } else if (/^blob\:/i.test(img.src)) { // Object URL
	        fileReader.onload = function(e) {
	          handleBinaryFile(e.target.result);
	        };
	        objectURLToBlob(img.src, function(blob) {
	          fileReader.readAsArrayBuffer(blob);
	        });
	      } else {
	        var http = new XMLHttpRequest();
	        http.onload = function() {
	          if (this.status == 200 || this.status === 0) {
	            handleBinaryFile(http.response);
	          } else {
	            throw "Could not load image";
	          }
	          http = null;
	        };
	        http.open("GET", img.src, true);
	        http.responseType = "arraybuffer";
	        http.send(null);
	      }
	    } else if (window.FileReader && (img instanceof window.Blob || img instanceof window.File)) {
	      fileReader.onload = function(e) {
	        if (debug) console.log("Got file of length " + e.target.result.byteLength);
	        handleBinaryFile(e.target.result);
	      };

	      fileReader.readAsArrayBuffer(img);
	    }
	  }

	  function findEXIFinJPEG(file) {
	    var dataView = new DataView(file);

	    if (debug) console.log("Got file of length " + file.byteLength);
	    if ((dataView.getUint8(0) != 0xFF) || (dataView.getUint8(1) != 0xD8)) {
	      if (debug) console.log("Not a valid JPEG");
	      return false; // not a valid jpeg
	    }

	    var offset = 2,
	      length = file.byteLength,
	      marker;

	    while (offset < length) {
	      if (dataView.getUint8(offset) != 0xFF) {
	        if (debug) console.log("Not a valid marker at offset " + offset + ", found: " + dataView.getUint8(offset));
	        return false; // not a valid marker, something is wrong
	      }

	      marker = dataView.getUint8(offset + 1);
	      if (debug) console.log(marker);

	      // we could implement handling for other markers here,
	      // but we're only looking for 0xFFE1 for EXIF data

	      if (marker == 225) {
	        if (debug) console.log("Found 0xFFE1 marker");

	        return readEXIFData(dataView, offset + 4, dataView.getUint16(offset + 2) - 2);

	        // offset += 2 + file.getShortAt(offset+2, true);

	      } else {
	        offset += 2 + dataView.getUint16(offset + 2);
	      }

	    }

	  }

	  function findIPTCinJPEG(file) {
	    var dataView = new DataView(file);

	    if (debug) console.log("Got file of length " + file.byteLength);
	    if ((dataView.getUint8(0) != 0xFF) || (dataView.getUint8(1) != 0xD8)) {
	      if (debug) console.log("Not a valid JPEG");
	      return false; // not a valid jpeg
	    }

	    var offset = 2,
	      length = file.byteLength;


	    var isFieldSegmentStart = function(dataView, offset) {
	      return (
	        dataView.getUint8(offset) === 0x38 &&
	        dataView.getUint8(offset + 1) === 0x42 &&
	        dataView.getUint8(offset + 2) === 0x49 &&
	        dataView.getUint8(offset + 3) === 0x4D &&
	        dataView.getUint8(offset + 4) === 0x04 &&
	        dataView.getUint8(offset + 5) === 0x04
	      );
	    };

	    while (offset < length) {

	      if (isFieldSegmentStart(dataView, offset)) {

	        // Get the length of the name header (which is padded to an even number of bytes)
	        var nameHeaderLength = dataView.getUint8(offset + 7);
	        if (nameHeaderLength % 2 !== 0) nameHeaderLength += 1;
	        // Check for pre photoshop 6 format
	        if (nameHeaderLength === 0) {
	          // Always 4
	          nameHeaderLength = 4;
	        }

	        var startOffset = offset + 8 + nameHeaderLength;
	        var sectionLength = dataView.getUint16(offset + 6 + nameHeaderLength);

	        return readIPTCData(file, startOffset, sectionLength);
	      }


	      // Not the marker, continue searching
	      offset++;

	    }

	  }
	  var IptcFieldMap = {
	    0x78: 'caption',
	    0x6E: 'credit',
	    0x19: 'keywords',
	    0x37: 'dateCreated',
	    0x50: 'byline',
	    0x55: 'bylineTitle',
	    0x7A: 'captionWriter',
	    0x69: 'headline',
	    0x74: 'copyright',
	    0x0F: 'category'
	  };

	  function readIPTCData(file, startOffset, sectionLength) {
	    var dataView = new DataView(file);
	    var data = {};
	    var fieldValue, fieldName, dataSize, segmentType, segmentSize;
	    var segmentStartPos = startOffset;
	    while (segmentStartPos < startOffset + sectionLength) {
	      if (dataView.getUint8(segmentStartPos) === 0x1C && dataView.getUint8(segmentStartPos + 1) === 0x02) {
	        segmentType = dataView.getUint8(segmentStartPos + 2);
	        if (segmentType in IptcFieldMap) {
	          dataSize = dataView.getInt16(segmentStartPos + 3);
	          segmentSize = dataSize + 5;
	          fieldName = IptcFieldMap[segmentType];
	          fieldValue = getStringFromDB(dataView, segmentStartPos + 5, dataSize);
	          // Check if we already stored a value with this name
	          if (data.hasOwnProperty(fieldName)) {
	            // Value already stored with this name, create multivalue field
	            if (data[fieldName] instanceof Array) {
	              data[fieldName].push(fieldValue);
	            } else {
	              data[fieldName] = [data[fieldName], fieldValue];
	            }
	          } else {
	            data[fieldName] = fieldValue;
	          }
	        }

	      }
	      segmentStartPos++;
	    }
	    return data;
	  }



	  function readTags(file, tiffStart, dirStart, strings, bigEnd) {
	    var entries = file.getUint16(dirStart, !bigEnd),
	      tags = {},
	      entryOffset, tag,
	      i;

	    for (i = 0; i < entries; i++) {
	      entryOffset = dirStart + i * 12 + 2;
	      tag = strings[file.getUint16(entryOffset, !bigEnd)];
	      if (!tag && debug) console.log("Unknown tag: " + file.getUint16(entryOffset, !bigEnd));
	      tags[tag] = readTagValue(file, entryOffset, tiffStart, dirStart, bigEnd);
	    }
	    return tags;
	  }


	  function readTagValue(file, entryOffset, tiffStart, dirStart, bigEnd) {
	    var type = file.getUint16(entryOffset + 2, !bigEnd),
	      numValues = file.getUint32(entryOffset + 4, !bigEnd),
	      valueOffset = file.getUint32(entryOffset + 8, !bigEnd) + tiffStart,
	      offset,
	      vals, val, n,
	      numerator, denominator;

	    switch (type) {
	      case 1: // byte, 8-bit unsigned int
	      case 7: // undefined, 8-bit byte, value depending on field
	        if (numValues == 1) {
	          return file.getUint8(entryOffset + 8, !bigEnd);
	        } else {
	          offset = numValues > 4 ? valueOffset : (entryOffset + 8);
	          vals = [];
	          for (n = 0; n < numValues; n++) {
	            vals[n] = file.getUint8(offset + n);
	          }
	          return vals;
	        }
	        break;
	      case 2: // ascii, 8-bit byte
	        offset = numValues > 4 ? valueOffset : (entryOffset + 8);
	        return getStringFromDB(file, offset, numValues - 1);

	      case 3: // short, 16 bit int
	        if (numValues == 1) {
	          return file.getUint16(entryOffset + 8, !bigEnd);
	        } else {
	          offset = numValues > 2 ? valueOffset : (entryOffset + 8);
	          vals = [];
	          for (n = 0; n < numValues; n++) {
	            vals[n] = file.getUint16(offset + 2 * n, !bigEnd);
	          }
	          return vals;
	        }
	        break;
	      case 4: // long, 32 bit int
	        if (numValues == 1) {
	          return file.getUint32(entryOffset + 8, !bigEnd);
	        } else {
	          vals = [];
	          for (n = 0; n < numValues; n++) {
	            vals[n] = file.getUint32(valueOffset + 4 * n, !bigEnd);
	          }
	          return vals;
	        }
	        break;
	      case 5: // rational = two long values, first is numerator, second is denominator
	        if (numValues == 1) {
	          numerator = file.getUint32(valueOffset, !bigEnd);
	          denominator = file.getUint32(valueOffset + 4, !bigEnd);
	          val = numerator / denominator;
	          val.numerator = numerator;
	          val.denominator = denominator;
	          return val;
	        } else {
	          vals = [];
	          for (n = 0; n < numValues; n++) {
	            numerator = file.getUint32(valueOffset + 8 * n, !bigEnd);
	            denominator = file.getUint32(valueOffset + 4 + 8 * n, !bigEnd);
	            vals[n] = numerator / denominator;
	            vals[n].numerator = numerator;
	            vals[n].denominator = denominator;
	          }
	          return vals;
	        }
	        break;
	      case 9: // slong, 32 bit signed int
	        if (numValues == 1) {
	          return file.getInt32(entryOffset + 8, !bigEnd);
	        } else {
	          vals = [];
	          for (n = 0; n < numValues; n++) {
	            vals[n] = file.getInt32(valueOffset + 4 * n, !bigEnd);
	          }
	          return vals;
	        }
	        break;
	      case 10: // signed rational, two slongs, first is numerator, second is denominator
	        if (numValues == 1) {
	          return file.getInt32(valueOffset, !bigEnd) / file.getInt32(valueOffset + 4, !bigEnd);
	        } else {
	          vals = [];
	          for (n = 0; n < numValues; n++) {
	            vals[n] = file.getInt32(valueOffset + 8 * n, !bigEnd) / file.getInt32(valueOffset + 4 + 8 * n, !bigEnd);
	          }
	          return vals;
	        }
	    }
	  }

	  function getStringFromDB(buffer, start, length) {
	    var outstr = "";
	    var n;
	    for (n = start; n < start + length; n++) {
	      outstr += String.fromCharCode(buffer.getUint8(n));
	    }
	    return outstr;
	  }

	  function readEXIFData(file, start) {
	    if (getStringFromDB(file, start, 4) != "Exif") {
	      if (debug) console.log("Not valid EXIF data! " + getStringFromDB(file, start, 4));
	      return false;
	    }

	    var bigEnd,
	      tags, tag,
	      exifData, gpsData,
	      tiffOffset = start + 6;

	    // test for TIFF validity and endianness
	    if (file.getUint16(tiffOffset) == 0x4949) {
	      bigEnd = false;
	    } else if (file.getUint16(tiffOffset) == 0x4D4D) {
	      bigEnd = true;
	    } else {
	      if (debug) console.log("Not valid TIFF data! (no 0x4949 or 0x4D4D)");
	      return false;
	    }

	    if (file.getUint16(tiffOffset + 2, !bigEnd) != 0x002A) {
	      if (debug) console.log("Not valid TIFF data! (no 0x002A)");
	      return false;
	    }

	    var firstIFDOffset = file.getUint32(tiffOffset + 4, !bigEnd);

	    if (firstIFDOffset < 0x00000008) {
	      if (debug) console.log("Not valid TIFF data! (First offset less than 8)", file.getUint32(tiffOffset + 4, !bigEnd));
	      return false;
	    }

	    tags = readTags(file, tiffOffset, tiffOffset + firstIFDOffset, TiffTags, bigEnd);

	    if (tags.ExifIFDPointer) {
	      exifData = readTags(file, tiffOffset, tiffOffset + tags.ExifIFDPointer, ExifTags, bigEnd);
	      for (tag in exifData) {
	        switch (tag) {
	          case "LightSource":
	          case "Flash":
	          case "MeteringMode":
	          case "ExposureProgram":
	          case "SensingMethod":
	          case "SceneCaptureType":
	          case "SceneType":
	          case "CustomRendered":
	          case "WhiteBalance":
	          case "GainControl":
	          case "Contrast":
	          case "Saturation":
	          case "Sharpness":
	          case "SubjectDistanceRange":
	          case "FileSource":
	            exifData[tag] = StringValues[tag][exifData[tag]];
	            break;

	          case "ExifVersion":
	          case "FlashpixVersion":
	            exifData[tag] = String.fromCharCode(exifData[tag][0], exifData[tag][1], exifData[tag][2], exifData[tag][3]);
	            break;

	          case "ComponentsConfiguration":
	            exifData[tag] =
	              StringValues.Components[exifData[tag][0]] +
	              StringValues.Components[exifData[tag][1]] +
	              StringValues.Components[exifData[tag][2]] +
	              StringValues.Components[exifData[tag][3]];
	            break;
	        }
	        tags[tag] = exifData[tag];
	      }
	    }

	    if (tags.GPSInfoIFDPointer) {
	      gpsData = readTags(file, tiffOffset, tiffOffset + tags.GPSInfoIFDPointer, GPSTags, bigEnd);
	      for (tag in gpsData) {
	        switch (tag) {
	          case "GPSVersionID":
	            gpsData[tag] = gpsData[tag][0] +
	              "." + gpsData[tag][1] +
	              "." + gpsData[tag][2] +
	              "." + gpsData[tag][3];
	            break;
	        }
	        tags[tag] = gpsData[tag];
	      }
	    }

	    return tags;
	  }

	  EXIF.getData = function(img, callback) {
	    if ((img instanceof Image || img instanceof HTMLImageElement) && !img.complete) return false;
	    if (!imageHasData(img)) {
	      getImageData(img, callback);
	    } else {
	      if (callback) {
	        callback(img);
	      }
	    }
	    return true;
	  };

	  EXIF.getTag = function(img, tag) {
	    if (!imageHasData(img)) return;
	    return img.exifdata[tag];
	  };

	  EXIF.getAllTags = function(img) {
	    if (!imageHasData(img)) return {};
	    var a,
	      data = img.exifdata,
	      tags = {};
	    for (a in data) {
	      if (data.hasOwnProperty(a)) {
	        tags[a] = data[a];
	      }
	    }
	    return tags;
	  };

	  EXIF.pretty = function(img) {
	    if (!imageHasData(img)) return "";
	    var a,
	      data = img.exifdata,
	      strPretty = "";
	    for (a in data) {
	      if (data.hasOwnProperty(a)) {
	        if (typeof data[a] == "object") {
	          if (data[a] instanceof Number) {
	            strPretty += a + " : " + data[a] + " [" + data[a].numerator + "/" + data[a].denominator + "]\r\n";
	          } else {
	            strPretty += a + " : [" + data[a].length + " values]\r\n";
	          }
	        } else {
	          strPretty += a + " : " + data[a] + "\r\n";
	        }
	      }
	    }
	    return strPretty;
	  };

	  EXIF.readFromBinaryFile = function(file) {
	    return findEXIFinJPEG(file);
	  };

	  if (true) {
	    !(__WEBPACK_AMD_DEFINE_ARRAY__ = [], __WEBPACK_AMD_DEFINE_RESULT__ = function() {
	      return EXIF;
	    }.apply(exports, __WEBPACK_AMD_DEFINE_ARRAY__), __WEBPACK_AMD_DEFINE_RESULT__ !== undefined && (module.exports = __WEBPACK_AMD_DEFINE_RESULT__));
	  }
	}.call(this));


/***/ },
/* 3 */
/***/ function(module, exports, __webpack_require__) {

	/* --------------------------------------------------
	Javascript Only Barcode_Reader (BarcodeReader) V1.6 by Eddie Larsson <https://github.com/EddieLa/BarcodeReader>

	This software is provided under the MIT license, http://opensource.org/licenses/MIT.
	All use of this software must include this
	text, including the reference to the creator of the original source code. The
	originator accepts no responsibility of any kind pertaining to
	use of this software.

	Copyright (c) 2013 Eddie Larsson

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in
	all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	THE SOFTWARE.

	------------------------ */

	var decoderWorkerBlob = function decoderWorkerBlob(){

	  function Rotate(data, width, height, rotation) {
	    var newData = [];
	    var x, y;
	    switch (rotation) {
	      case 90:
	        for (x = 0; x < width * 4; x += 4) {
	          for (y = width * 4 * (height - 1); y >= 0; y -= width * 4) {
	            newData.push(data[x + y]);
	            newData.push(data[x + y + 1]);
	            newData.push(data[x + y + 2]);
	            newData.push(data[x + y + 3]);
	          }
	        }
	        break;
	      case -90:
	        for (x = width * 4 - 4; x >= 0; x -= 4) {
	          for (y = 0; y < data.length; y += width * 4) {
	            newData.push(data[x + y]);
	            newData.push(data[x + y + 1]);
	            newData.push(data[x + y + 2]);
	            newData.push(data[x + y + 3]);
	          }
	        }
	        break;
	      case 180:
	        for (y = width * 4 * (height - 1); y >= 0; y -= width * 4) {
	          for (x = width * 4 - 4; x >= 0; x -= 4) {
	            newData.push(data[x + y]);
	            newData.push(data[x + y + 1]);
	            newData.push(data[x + y + 2]);
	            newData.push(data[x + y + 3]);
	          }
	        }
	    }
	    return new Uint8ClampedArray(newData);
	  }

	  function BoxFilter(data, width, radius) {
	    var elements = [];
	    var sum = [];
	    var val;
	    var x, y, i;
	    for (x = 0; x < width; x++) {
	      elements.push([]);
	      sum.push(0);
	      for (y = 0; y < (radius + 1) * width; y += width) {
	        elements[elements.length - 1].push(data[x + y]);
	        sum[sum.length - 1] = sum[sum.length - 1] + data[x + y];
	      }
	    }
	    var newData = [];
	    for (y = 0; y < data.length; y += width) {
	      for (x = 0; x < width; x++) {
	        var newVal = 0;
	        var length = 0;
	        for (i = x; i >= 0; i--) {
	          newVal += sum[i];
	          length++;
	          if (length === radius + 1) break;
	        }
	        var tempLength = 0;
	        for (i = x + 1; i < width; i++) {
	          newVal += sum[i];
	          length++;
	          tempLength++;
	          if (tempLength === radius) break;
	        }
	        length *= elements[0].length;
	        newVal /= length;
	        newData.push(newVal);
	      }
	      if (y - radius * width >= 0) {
	        for (i = 0; i < elements.length; i++) {
	          val = elements[i].shift();
	          sum[i] = sum[i] - val;
	        }
	      }
	      if (y + (radius + 1) * width < data.length) {
	        for (i = 0; i < elements.length; i++) {
	          val = data[i + y + (radius + 1) * width];
	          elements[i].push(val);
	          sum[i] = sum[i] + val;
	        }
	      }
	    }
	    return newData;
	  }

	  function Scale(data, width, height) {
	    var newData = [];
	    var x, y;
	    for (y = 0; y < data.length; y += width * 8) {
	      for (x = 0; x < width * 4; x += 8) {
	        var r = (data[y + x] + data[y + x + 4] + data[y + width * 4 + x] + data[y + width * 4 + x + 4]) / 4;
	        newData.push(r);
	        var g = (data[y + x + 1] + data[y + x + 4 + 1] + data[y + width * 4 + x + 1] + data[y + width * 4 + x + 4 + 1]) / 4;
	        newData.push(g);
	        var b = (data[y + x + 2] + data[y + x + 4 + 2] + data[y + width * 4 + x + 2] + data[y + width * 4 + x + 4 + 2]) / 4;
	        newData.push(b);
	        newData.push(255);
	      }
	    }
	    return new Uint8ClampedArray(newData);
	  }

	  function IntensityGradient(data, width) {
	    var newData = [];
	    var max = Number.MIN_VALUE;
	    var min = Number.MAX_VALUE;
	    var x, y, i;
	    for (y = 0; y < data.length; y += width * 4) {
	      for (x = 0; x < width * 4; x += 4) {
	        var horizontalDiff = 0;
	        var verticalDiff = 0;
	        for (i = 1; i < 2; i++) {
	          if (x + i * 4 < width * 4) {
	            horizontalDiff = horizontalDiff + Math.abs(data[y + x] - data[y + x + i * 4]);
	          }
	          if (y + width * 4 * i < data.length) {
	            verticalDiff += verticalDiff + Math.abs(data[y + x] - data[y + x + width * 4 * i]);
	          }
	        }
	        var diff = horizontalDiff - verticalDiff;
	        max = diff > max ? diff : max;
	        min = diff < min ? diff : min;
	        newData.push(diff);
	      }
	    }
	    if (min < 0) {
	      for (i = 0; i < newData.length; i++) {
	        newData[i] = newData[i] - min;
	      }
	      min = 0;
	    }
	    return newData;
	  }

	  function greyScale(data) {
	    var i;
	    for (i = 0; i < data.length; i += 4) {
	      var max = 0;
	      var min = 255;
	      max = data[i] > max ? data[i] : max;
	      max = data[i + 1] > max ? data[i + 1] : max;
	      max = data[i + 2] > max ? data[i + 2] : max;
	      min = data[i] < min ? data[i] : min;
	      min = data[i + 1] < min ? data[i + 1] : min;
	      min = data[i + 2] < min ? data[i + 2] : min;
	      data[i] = data[i + 1] = data[i + 2] = (max + min) / 2;
	    }
	  }

	  function histogram(data) {
	    var i;
	    var hist = [];
	    for (i = 0; i < 256; i++) {
	      hist[i] = 0;
	    }
	    for (i = 0; i < data.length; i += 4) {
	      hist[data[i]] = hist[data[i]] + 1;
	    }
	    return hist;
	  }

	  function otsu(histogram, total) {
	    var i;
	    var sum = 0;
	    for (i = 1; i < histogram.length; ++i)
	      sum += i * histogram[i];
	    var sumB = 0;
	    var wB = 0;
	    var wF = 0;
	    var mB;
	    var mF;
	    var max = 0.0;
	    var between = 0.0;
	    var threshold1 = 0.0;
	    var threshold2 = 0.0;
	    for (i = 0; i < histogram.length; ++i) {
	      wB += histogram[i];
	      if (wB === 0)
	        continue;
	      wF = total - wB;
	      if (wF === 0)
	        break;
	      sumB += i * histogram[i];
	      mB = sumB / wB;
	      mF = (sum - sumB) / wF;
	      between = wB * wF * Math.pow(mB - mF, 2);
	      if (between >= max) {
	        threshold1 = i;
	        if (between > max) {
	          threshold2 = i;
	        }
	        max = between;
	      }
	    }
	    return (threshold1 + threshold2) / 2.0;
	  }

	  function CreateImageData() {
	    Image.data = new Uint8ClampedArray(Image.width * Image.height * 4);
	    var Converter;
	    var x, y;
	    for (y = 0; y < Image.height; y++) {
	      for (x = 0; x < Image.width; x++) {
	        Converter = y * 4 * Image.width;
	        Image.data[Converter + x * 4] = Image.table[x][y][0];
	        Image.data[Converter + x * 4 + 1] = Image.table[x][y][1];
	        Image.data[Converter + x * 4 + 2] = Image.table[x][y][2];
	        Image.data[Converter + x * 4 + 3] = Image.table[x][y][3];
	      }
	    }
	  }

	  function CreateScanImageData() {
	    ScanImage.data = new Uint8ClampedArray(ScanImage.width * ScanImage.height * 4);
	    var Converter;
	    var x, y;
	    for (y = 0; y < ScanImage.height; y++) {
	      for (x = 0; x < ScanImage.width; x++) {
	        Converter = y * 4 * ScanImage.width;
	        ScanImage.data[Converter + x * 4] = ScanImage.table[x][y][0];
	        ScanImage.data[Converter + x * 4 + 1] = ScanImage.table[x][y][1];
	        ScanImage.data[Converter + x * 4 + 2] = ScanImage.table[x][y][2];
	        ScanImage.data[Converter + x * 4 + 3] = ScanImage.table[x][y][3];
	      }
	    }
	  }

	  function CreateTable() {
	    Image.table = [];
	    var tempArray = [];
	    var i, j;
	    for (i = 0; i < Image.width * 4; i += 4) {
	      tempArray = [];
	      for (j = i; j < Image.data.length; j += Image.width * 4) {
	        tempArray.push([Image.data[j], Image.data[j + 1], Image.data[j + 2], Image.data[j + 3]]);
	      }
	      Image.table.push(tempArray);
	    }
	  }

	  function CreateScanTable() {
	    ScanImage.table = [];
	    var tempArray = [];
	    var i, j;
	    for (i = 0; i < ScanImage.width * 4; i += 4) {
	      tempArray = [];
	      for (j = i; j < ScanImage.data.length; j += ScanImage.width * 4) {
	        tempArray.push([ScanImage.data[j], ScanImage.data[j + 1], ScanImage.data[j + 2], ScanImage.data[j + 3]]);
	      }
	      ScanImage.table.push(tempArray);
	    }
	  }

	  function EnlargeTable(h, w) {
	    var TempArray = [];
	    var x, y, i;
	    for (x = 0; x < Image.width; x++) {
	      TempArray = [];
	      for (y = 0; y < Image.height; y++) {
	        for (i = 0; i < h; i++) {
	          TempArray.push(Image.table[x][y]);
	        }
	      }
	      Image.table[x] = TempArray.slice();
	    }
	    TempArray = Image.table.slice();
	    for (x = 0; x < Image.width; x++) {
	      for (i = 0; i < w; i++) {
	        Image.table[x * w + i] = TempArray[x].slice();
	      }
	    }
	    Image.width = Image.table.length;
	    Image.height = Image.table[0].length;
	    CreateImageData();
	  }

	  function ScaleHeight(scale) {
	    var tempArray = [];
	    var avrgRed = 0;
	    var avrgGreen = 0;
	    var avrgBlue = 0;
	    var i, j, k;
	    for (i = 0; i < Image.height - scale; i += scale) {
	      for (j = 0; j < Image.width; j++) {
	        avrgRed = 0;
	        avrgGreen = 0;
	        avrgBlue = 0;
	        for (k = i; k < i + scale; k++) {
	          avrgRed += Image.table[j][k][0];
	          avrgGreen += Image.table[j][k][1];
	          avrgBlue += Image.table[j][k][2];
	        }
	        tempArray.push(avrgRed / scale);
	        tempArray.push(avrgGreen / scale);
	        tempArray.push(avrgBlue / scale);
	        tempArray.push(255);
	      }
	    }
	    return new Uint8ClampedArray(tempArray);
	  }

	  function Intersects(rectOne, rectTwo) {
	    return (rectOne[0][0] <= rectTwo[0][1] &&
	      rectTwo[0][0] <= rectOne[0][1] &&
	      rectOne[1][0] <= rectTwo[1][1] &&
	      rectTwo[1][0] <= rectOne[1][1]);
	  }

	  function maxLocalization(max, maxPos, data) {
	    var originalMax = max;
	    var rects = [];
	    var x, y, i;
	    do {
	      var startX = maxPos % Image.width;
	      var startY = (maxPos - startX) / Image.width;
	      var minY = 0;
	      var maxY = Image.height;
	      var minX = 0;
	      var maxX = Image.width - 1;
	      for (y = startY; y < Image.height - 1; y++) {
	        if (Image.table[startX][y + 1][0] === 0) {
	          maxY = y;
	          break;
	        }
	      }
	      for (y = startY; y > 0; y--) {
	        if (Image.table[startX][y - 1][0] === 0) {
	          minY = y;
	          break;
	        }
	      }
	      for (x = startX; x < Image.width - 1; x++) {
	        if (Image.table[x + 1][startY][0] === 0) {
	          maxX = x;
	          break;
	        }
	      }
	      for (x = startX; x > 0; x--) {
	        if (Image.table[x - 1][startY][0] === 0) {
	          minX = x;
	          break;
	        }
	      }
	      for (y = minY * Image.width; y <= maxY * Image.width; y += Image.width) {
	        for (x = minX; x <= maxX; x++) {
	          data[y + x] = 0;
	        }
	      }
	      var newRect = [
	        [minX, maxX],
	        [minY, maxY]
	      ];
	      for (i = 0; i < rects.length; i++) {
	        if (Intersects(newRect, rects[i])) {
	          if (rects[i][0][1] - rects[i][0][0] > newRect[0][1] - newRect[0][0]) {
	            rects[i][0][0] = rects[i][0][0] < newRect[0][0] ? rects[i][0][0] : newRect[0][0];
	            rects[i][0][1] = rects[i][0][1] > newRect[0][1] ? rects[i][0][1] : newRect[0][1];
	            newRect = [];
	            break;
	          } else {
	            rects[i][0][0] = rects[i][0][0] < newRect[0][0] ? rects[i][0][0] : newRect[0][0];
	            rects[i][0][1] = rects[i][0][1] > newRect[0][1] ? rects[i][0][1] : newRect[0][1];
	            rects[i][1][0] = newRect[1][0];
	            rects[i][1][1] = newRect[1][1];
	            newRect = [];
	            break;
	          }
	        }
	      }
	      if (newRect.length > 0) {
	        rects.push(newRect);
	      }
	      max = 0;
	      maxPos = 0;
	      var newMaxPos = 0;
	      for (i = 0; i < data.length; i++) {
	        if (data[i] > max) {
	          max = data[i];
	          maxPos = i;
	        }
	      }
	    } while (max > originalMax * 0.70);
	    return rects;
	  }

	  function ImgProcessing() {
	    greyScale(Image.data);
	    var newData = IntensityGradient(Image.data, Image.width);
	    newData = BoxFilter(newData, Image.width, 15);
	    var min = newData[0];
	    var i, x, y;
	    for (i = 1; i < newData.length; i++) {
	      min = min > newData[i] ? newData[i] : min;
	    }
	    var max = 0;
	    var maxPos = 0;
	    var avrgLight = 0;
	    for (i = 0; i < newData.length; i++) {
	      newData[i] = Math.round((newData[i] - min));
	      avrgLight += newData[i];
	      if (max < newData[i]) {
	        max = newData[i];
	        maxPos = i;
	      }
	    }
	    avrgLight /= newData.length;
	    if (avrgLight < 15) {
	      newData = BoxFilter(newData, Image.width, 8);
	      min = newData[0];
	      for (i = 1; i < newData.length; i++) {
	        min = min > newData[i] ? newData[i] : min;
	      }
	      max = 0;
	      maxPos = 0;
	      for (i = 0; i < newData.length; i++) {
	        newData[i] = Math.round((newData[i] - min));
	        if (max < newData[i]) {
	          max = newData[i];
	          maxPos = i;
	        }
	      }
	    }
	    var hist = [];
	    for (i = 0; i <= max; i++) {
	      hist[i] = 0;
	    }
	    for (i = 0; i < newData.length; i++) {
	      hist[newData[i]] = hist[newData[i]] + 1;
	    }
	    var thresh = otsu(hist, newData.length);
	    for (i = 0; i < newData.length; i++) {
	      if (newData[i] < thresh) {
	        Image.data[i * 4] = Image.data[i * 4 + 1] = Image.data[i * 4 + 2] = 0;
	      } else {
	        Image.data[i * 4] = Image.data[i * 4 + 1] = Image.data[i * 4 + 2] = 255;
	      }
	    }
	    CreateTable();
	    var rects = maxLocalization(max, maxPos, newData);
	    var feedBack = [];
	    for (i = 0; i < rects.length; i++) {
	      feedBack.push({
	        x: rects[i][0][0],
	        y: rects[i][1][0],
	        width: rects[i][0][1] - rects[i][0][0],
	        height: rects[i][1][1] - rects[i][1][0]
	      });
	    }
	    if (feedBack.length > 0) postMessage({
	      result: feedBack,
	      success: "localization"
	    });
	    allTables = [];
	    for (i = 0; i < rects.length; i++) {
	      var newTable = [];
	      for (x = rects[i][0][0] * 2; x < rects[i][0][1] * 2; x++) {
	        var tempArray = [];
	        for (y = rects[i][1][0] * 2; y < rects[i][1][1] * 2; y++) {
	          tempArray.push([ScanImage.table[x][y][0], ScanImage.table[x][y][1], ScanImage.table[x][y][2], 255]);
	        }
	        newTable.push(tempArray);
	      }
	      if (newTable.length < 1) continue;
	      Image.table = newTable;
	      Image.width = newTable.length;
	      Image.height = newTable[0].length;
	      CreateImageData();
	      allTables.push({
	        table: newTable,
	        data: new Uint8ClampedArray(Image.data),
	        width: Image.width,
	        height: Image.height
	      });
	    }
	  }

	  function showImage(data, width, height) {
	    postMessage({
	      result: data,
	      width: width,
	      height: height,
	      success: "image"
	    });
	  }

	  function Main() {
	    ImgProcessing();
	    var allResults = [];
	    var tempObj;
	    var tempData;
	    var hist;
	    var val;
	    var thresh;
	    var start;
	    var end;
	    var z, i;
	    for (z = 0; z < allTables.length; z++) {
	      Image = allTables[z];
	      var scaled = ScaleHeight(30);
	      var variationData;
	      var incrmt = 0;
	      var format = "";
	      var first = true;
	      var eanStatistics = {};
	      var eanOrder = [];
	      Selection = false;
	      do {
	        tempData = scaled.subarray(incrmt, incrmt + Image.width * 4);
	        hist = [];
	        for (i = 0; i < 256; i++) {
	          hist[i] = 0;
	        }
	        for (i = 0; i < tempData.length; i += 4) {
	          val = Math.round((tempData[i] + tempData[i + 1] + tempData[i + 2]) / 3);
	          hist[val] = hist[val] + 1;
	        }
	        thresh = otsu(hist, tempData.length / 4);
	        start = thresh < 41 ? 1 : thresh - 40;
	        end = thresh > 254 - 40 ? 254 : thresh + 40;
	        variationData = yStraighten(tempData, start, end);
	        Selection = BinaryString(variationData);
	        if (Selection.string) {
	          format = Selection.format;
	          tempObj = Selection;
	          Selection = Selection.string;
	          if (format === "EAN-13") {
	            if (typeof eanStatistics[Selection] === 'undefined') {
	              eanStatistics[Selection] = {
	                count: 1,
	                correction: tempObj.correction
	              };
	              eanOrder.push(Selection);
	            } else {
	              eanStatistics[Selection].count = eanStatistics[Selection].count + 1;
	              eanStatistics[Selection].correction = eanStatistics[Selection].correction + tempObj.correction;
	            }
	            Selection = false;
	          }
	        } else {
	          Selection = false;
	        }
	        incrmt += Image.width * 4;
	      } while (!Selection && incrmt < scaled.length);
	      if (Selection && format !== "EAN-13") allResults.push({
	        Format: format,
	        Value: Selection
	      });
	      if (format === "EAN-13") Selection = false;
	      if (!Selection) {
	        EnlargeTable(4, 2);
	        incrmt = 0;
	        scaled = ScaleHeight(20);
	        do {
	          tempData = scaled.subarray(incrmt, incrmt + Image.width * 4);
	          hist = [];
	          for (i = 0; i < 256; i++) {
	            hist[i] = 0;
	          }
	          for (i = 0; i < tempData.length; i += 4) {
	            val = Math.round((tempData[i] + tempData[i + 1] + tempData[i + 2]) / 3);
	            hist[val] = hist[val] + 1;
	          }
	          thresh = otsu(hist, tempData.length / 4);
	          start = thresh < 40 ? 0 : thresh - 40;
	          end = thresh > 255 - 40 ? 255 : thresh + 40;
	          variationData = yStraighten(tempData, start, end);
	          Selection = BinaryString(variationData);
	          if (Selection.string) {
	            format = Selection.format;
	            tempObj = Selection;
	            Selection = Selection.string;
	            if (format === "EAN-13") {
	              if (typeof eanStatistics[Selection] === 'undefined') {
	                eanStatistics[Selection] = {
	                  count: 1,
	                  correction: tempObj.correction
	                };
	                eanOrder.push(Selection);
	              } else {
	                eanStatistics[Selection].count = eanStatistics[Selection].count + 1;
	                eanStatistics[Selection].correction = eanStatistics[Selection].correction + tempObj.correction;
	              }
	              Selection = false;
	            }
	          } else {
	            Selection = false;
	          }
	          incrmt += Image.width * 4;
	        } while (!Selection && incrmt < scaled.length);
	        if (format === "EAN-13") {
	          var points = {};
	          for (var key in eanStatistics) {
	            eanStatistics[key].correction = eanStatistics[key].correction / eanStatistics[key].count;
	            var pointTemp = eanStatistics[key].correction;
	            pointTemp -= eanStatistics[key].count;
	            pointTemp += eanOrder.indexOf(key);
	            points[key] = pointTemp;
	          }
	          var minPoints = Number.POSITIVE_INFINITY;
	          var tempString = "";
	          for (var point in points) {
	            if (points[point] < minPoints) {
	              minPoints = points[point];
	              tempString = key;
	            }
	          }
	          if (minPoints < 11) {
	            Selection = tempString;
	          } else {
	            Selection = false;
	          }
	        }
	        if (Selection) allResults.push({
	          Format: format,
	          Value: Selection
	        });
	      }
	      if (allResults.length > 0 && !Multiple) break;
	    }
	    return allResults;
	  }

	  function yStraighten(img, start, end) {
	    var average = 0;
	    var threshold;
	    var newImg = new Uint8ClampedArray(Image.width * (end - start + 1) * 4);
	    var i, j;
	    for (i = 0; i < newImg.length; i++) {
	      newImg[i] = 255;
	    }
	    for (i = 0; i < Image.width * 4; i += 4) {
	      threshold = end;
	      average = (img[i] + img[i + 1] + img[i + 2]) / 3;
	      if (i < Image.width * 4 - 4) {
	        average += (img[i + 4] + img[i + 5] + img[i + 6]) / 3;
	        average /= 2;
	      }
	      for (j = i; j < newImg.length; j += Image.width * 4) {
	        if (average < threshold) {
	          newImg[j] = newImg[j + 1] = newImg[j + 2] = 0;
	        }
	        threshold--;
	      }
	    }
	    return newImg;
	  }

	  function CheckEan13(values, middle) {
	    if (middle) {
	      if (values.length !== 5) return false;
	    } else {
	      if (values.length !== 3) return false;
	    }
	    var avrg = 0;
	    var i;
	    for (i = 0; i < values.length; i++) {
	      avrg += values[i];
	    }
	    avrg /= values.length;
	    for (i = 0; i < values.length; i++) {
	      if (values[i] / avrg < 0.5 || values[i] / avrg > 1.5) return false;
	    }
	    return true;
	  }

	  function TwoOfFiveStartEnd(values, start) {
	    if (values.length < 5 || values.length > 6) return false;
	    var maximum = 0;
	    var TwoOfFiveMax = [0, 0];
	    var u;
	    for (u = 0; u < values.length; u++) {
	      if (values[u] > maximum) {
	        maximum = values[u];
	        TwoOfFiveMax[0] = u;
	      }
	    }
	    maximum = 0;
	    for (u = 0; u < values.length; u++) {
	      if (u === TwoOfFiveMax[0]) continue;
	      if (values[u] > maximum) {
	        maximum = values[u];
	        TwoOfFiveMax[1] = u;
	      }
	    }
	    if (start) {
	      return TwoOfFiveMax[0] + TwoOfFiveMax[1] === 2;
	    } else {
	      return TwoOfFiveMax[0] + TwoOfFiveMax[1] === 2;
	    }
	  }

	  function CheckInterleaved(values, start) {
	    var average = 0;
	    var i;
	    for (i = 0; i < values.length; i++) {
	      average += values[i];
	    }
	    average /= 4;
	    if (start) {
	      if (values.length !== 4) return false;
	      for (i = 0; i < values.length; i++) {
	        if (values[i] / average < 0.5 || values[i] / average > 1.5) return false;
	      }
	      return true;
	    } else {
	      if (values.length !== 3) return false;
	      var max = 0;
	      var pos;
	      for (i = 0; i < values.length; i++) {
	        if (values[i] > max) {
	          max = values[i];
	          pos = i;
	        }
	      }
	      if (pos !== 0) return false;
	      if (values[0] / average < 1.5 || values[0] / average > 2.5) return false;
	      for (i = 1; i < values.length; i++) {
	        if (values[i] / average < 0.5 || values[i] / average > 1.5) return false;
	      }
	      return true;
	    }
	  }

	  function BinaryConfiguration(binaryString, type) {
	    var result = [];
	    var binTemp = [];
	    var count = 0;
	    var bars;
	    var len;
	    var totalBars;
	    var i;
	    if (type === "Code128" || type === "Code93") {
	      totalBars = 6;
	      len = binaryString[0];
	      if (type === "Code128") len /= 2;
	      for (i = 0; i < binaryString.length; i++) {
	        if (binaryString[i] > len * 6) {
	          binaryString.splice(i, binaryString.length);
	          break;
	        }
	      }
	      do {
	        if (binaryString.length === 7 && type === "Code128") {
	          result.push(binaryString.splice(0, binaryString.length));
	        } else {
	          result.push(binaryString.splice(0, totalBars));
	        }
	        if (type === "Code93" && binaryString.length < 6) binaryString.splice(0, totalBars);
	      } while (binaryString.length > 0);
	    }
	    if (type === "Code39") {
	      totalBars = 9;
	      len = binaryString[0];
	      for (i = 0; i < binaryString.length; i++) {
	        if (binaryString[i] > len * 5) {
	          binaryString.splice(i, binaryString.length);
	          break;
	        }
	      }
	      do {
	        result.push(binaryString.splice(0, totalBars));
	        binaryString.splice(0, 1);
	      } while (binaryString.length > 0);
	    }
	    if (type === "EAN-13") {
	      totalBars = 4;
	      len = binaryString[0];
	      var secureCount = 0;
	      for (i = 0; i < binaryString.length; i++) {
	        if (binaryString[i] > len * 6) {
	          binaryString.splice(i, binaryString.length);
	          break;
	        }
	      }
	      if (CheckEan13(binaryString.splice(0, 3), false)) secureCount++;
	      count = 0;
	      do {
	        result.push(binaryString.splice(0, totalBars));
	        count++;
	        if (count === 6)
	          if (CheckEan13(binaryString.splice(0, 5), true)) secureCount++;
	      } while (result.length < 12 && binaryString.length > 0);
	      if (CheckEan13(binaryString.splice(0, 3), false)) secureCount++;
	      if (secureCount < 2) return [];
	    }
	    if (type === "2Of5") {
	      totalBars = 5;
	      len = binaryString[0] / 2;
	      for (i = 0; i < binaryString.length; i++) {
	        if (binaryString[i] > len * 5) {
	          binaryString.splice(i, binaryString.length);
	          break;
	        }
	      }
	      var temp = binaryString.splice(0, 6);
	      result.push(temp);
	      do {
	        binTemp = [];
	        for (i = 0; i < totalBars; i++) {
	          binTemp.push(binaryString.splice(0, 1)[0]);
	          // binaryString.splice(0, 1)[0];
	        }
	        result.push(binTemp);
	        if (binaryString.length === 5) result.push(binaryString.splice(0, 5));
	      } while (binaryString.length > 0);
	    }
	    if (type === "Inter2Of5") {
	      totalBars = 5;
	      len = binaryString[0];
	      for (i = 0; i < binaryString.length; i++) {
	        if (binaryString[i] > len * 5) {
	          binaryString.splice(i, binaryString.length);
	          break;
	        }
	      }
	      result.push(binaryString.splice(0, 4));
	      var binTempWhite = [];
	      do {
	        binTemp = [];
	        binTempWhite = [];
	        for (i = 0; i < totalBars; i++) {
	          binTemp.push(binaryString.splice(0, 1)[0]);
	          binTempWhite.push(binaryString.splice(0, 1)[0]);
	        }
	        result.push(binTemp);
	        result.push(binTempWhite);
	        if (binaryString.length === 3) result.push(binaryString.splice(0, 3));
	      } while (binaryString.length > 0);
	    }
	    if (type === "Codabar") {
	      totalBars = 7;
	      len = binaryString[0];
	      for (i = 0; i < binaryString.length; i++) {
	        if (binaryString[i] > len * 5) {
	          binaryString.splice(i, binaryString.length);
	          break;
	        }
	      }
	      do {
	        result.push(binaryString.splice(0, totalBars));
	        binaryString.splice(0, 1);
	      } while (binaryString.length > 0);
	    }
	    return result;
	  }

	  function BinaryString(img, type) {
	    var binaryString = [];
	    var binTemp = [];
	    var container = 255;
	    var count = 0;
	    var format;
	    var tempString;
	    var j, i;
	    for (j = 0; j < img.length - Image.width * 4; j += Image.width * 4) {
	      var SlicedArray = img.subarray(j, j + Image.width * 4);
	      binaryString = [];
	      i = 0;
	      while (SlicedArray[i] === 255) {
	        i += 4;
	      }
	      while (i < SlicedArray.length) {
	        count = 0;
	        container = SlicedArray[i];
	        while (SlicedArray[i] === container && i < SlicedArray.length) {
	          count++;
	          i += 4;
	        }
	        binaryString.push(count);
	      }
	      if (binaryString.length > 2 && binaryString[0] <= binaryString[1] / 10) {
	        binaryString.splice(0, 2);
	      }
	      var binaryHolder = binaryString.slice();
	      var success = false;
	      for (i = 0; i < FormatPriority.length; i++) {
	        binaryString = binaryHolder.slice();
	        var first;
	        var second;
	        binaryString = BinaryConfiguration(binaryString, FormatPriority[i]);
	        if (FormatPriority[i] === "2Of5" || FormatPriority[i] === "Inter2Of5") {
	          first = binaryString.splice(0, 1)[0];
	          second = binaryString.splice(binaryString.length - 1, 1)[0];
	        }
	        binTemp = Distribution(binaryString, FormatPriority[i]);
	        if (FormatPriority[i] === "EAN-13") {
	          binaryString = binTemp.data;
	          corrections = binTemp.correction;
	        } else {
	          binaryString = binTemp;
	        }
	        if (typeof binaryString === 'undefined') continue;
	        if (binaryString.length > 4 || (FormatPriority[i] === "Code39" && binaryString.length > 2)) {
	          if (FormatPriority[i] === "Code128") {
	            if (CheckCode128(binaryString)) {
	              binaryString = DecodeCode128(binaryString);
	              success = true;
	            }
	          } else if (FormatPriority[i] === "Code93") {
	            if (CheckCode93(binaryString)) {
	              binaryString = DecodeCode93(binaryString);
	              success = true;
	            }
	          } else if (FormatPriority[i] === "Code39") {
	            if (CheckCode39(binaryString)) {
	              binaryString = DecodeCode39(binaryString);
	              success = true;
	            }
	          } else if (FormatPriority[i] === "EAN-13") {
	            tempString = DecodeEAN13(binaryString);
	            if (tempString) {
	              if (tempString.length === 13) {
	                binaryString = tempString;
	                success = true;
	              }
	            }
	          } else if (FormatPriority[i] === "2Of5" || FormatPriority[i] === "Inter2Of5") {
	            if (FormatPriority[i] === "2Of5") {
	              if (typeof first !== 'undefined')
	                if (!TwoOfFiveStartEnd(first, true)) continue;
	              if (typeof second !== 'undefined')
	                if (!TwoOfFiveStartEnd(second, false)) continue;
	            }
	            if (FormatPriority[i] === "Inter2Of5") {
	              if (typeof first !== 'undefined')
	                if (!CheckInterleaved(first, true)) continue;
	              if (typeof second !== 'undefined')
	                if (!CheckInterleaved(second, false)) continue;
	            }
	            tempString = Decode2Of5(binaryString);
	            if (tempString) {
	              binaryString = tempString;
	              success = true;
	            }
	          } else if (FormatPriority[i] === "Codabar") {
	            tempString = DecodeCodaBar(binaryString);
	            if (tempString) {
	              binaryString = tempString;
	              success = true;
	            }
	          }
	        }
	        if (success) {
	          format = FormatPriority[i];
	          if (format === "Inter2Of5") format = "Interleaved 2 of 5";
	          if (format === "2Of5") format = "Standard 2 of 5";
	          break;
	        }
	      }
	      if (success) break;
	    }
	    if (format === "Code128") {
	      if (typeof binaryString.string === 'string') {
	        return binaryString;
	      } else {
	        return {
	          string: false
	        };
	      }
	    }
	    if (typeof binaryString === 'string') {
	      if (format === "EAN-13") {
	        return {
	          string: binaryString,
	          format: format,
	          correction: corrections
	        };
	      } else {
	        return {
	          string: binaryString,
	          format: format
	        };
	      }
	    } else {
	      return {
	        string: false
	      };
	    }
	  }

	  function Distribution(totalBinArray, type) {
	    var testData = 0;
	    var result = [];
	    var totalBars;
	    var total;
	    var maxLength;
	    var k, i, j;
	    var blackMax;
	    var whiteMax;
	    var wideAvrg;
	    var narrowAvrg;
	    var prevPos;
	    var wideValues;
	    var max;

	    type = availableFormats.indexOf(type);

	    if (type === 0) {
	      total = 11;
	      totalBars = 6;
	      maxLength = 4;
	    } else if (type === 1) {
	      total = 9;
	      totalBars = 6;
	      maxLength = 4;
	    } else if (type === 2) {
	      total = 12;
	      totalBars = 9;
	    } else if (type === 3) {
	      total = 7;
	      totalBars = 4;
	      maxLength = 4;
	    } else if (type === 6) {
	      totalBars = 7;
	    }
	    for (k = 0; k < totalBinArray.length; k++) {
	      var BinArray = totalBinArray[k];
	      var sum = 0;
	      var counter = 0;
	      var tempBin = [];
	      var narrowArr = [];
	      var wideArr = [];
	      if (type === 6) {
	        var upperTolerance = 1.5;
	        var lowerTolerance = 1 / 2;
	        if (BinArray.length !== 7) return [];
	        if (k === 0 || k === totalBinArray.length - 1) {
	          whiteMax = [
	            [0, 0],
	            [0, 0]
	          ];
	          blackMax = [0, 0];
	          for (i = 0; i < BinArray.length; i++) {
	            if (i % 2 === 0) {
	              if (BinArray[i] > blackMax[0]) {
	                blackMax[0] = BinArray[i];
	                blackMax[1] = i;
	              }
	            } else {
	              if (BinArray[i] > whiteMax[0][0]) {
	                whiteMax[0][0] = BinArray[i];
	                prevPos = whiteMax[0][1];
	                whiteMax[0][1] = i;
	                i = prevPos - 1;
	                continue;
	              }
	              if (BinArray[i] > whiteMax[1][0] && i !== whiteMax[0][1]) {
	                whiteMax[1][0] = BinArray[i];
	                whiteMax[1][1] = i;
	              }
	            }
	          }
	          if (SecureCodabar) {
	            wideAvrg = whiteMax[0][0] + whiteMax[1][0] + blackMax[0];
	            wideAvrg /= 3;
	            wideValues = [whiteMax[0][0], whiteMax[1][0], blackMax[0]];
	            for (i = 0; i < wideValues.length; i++) {
	              if (wideValues[i] / wideAvrg > upperTolerance || wideValues[i] / wideAvrg < lowerTolerance) return [];
	            }
	            narrowAvrg = 0;
	            for (i = 0; i < BinArray.length; i++) {
	              if (i === blackMax[1] || i === whiteMax[0][1] || i === whiteMax[1][1]) continue;
	              narrowAvrg += BinArray[i];
	            }
	            narrowAvrg /= 4;
	            for (i = 0; i < BinArray.length; i++) {
	              if (i === blackMax[1] || i === whiteMax[0][1] || i === whiteMax[1][1]) continue;
	              if (BinArray[i] / narrowAvrg > upperTolerance || BinArray[i] / narrowAvrg < lowerTolerance) return [];
	            }
	          }
	          for (i = 0; i < BinArray.length; i++) {
	            if (i === blackMax[1] || i === whiteMax[0][1] || i === whiteMax[1][1]) {
	              tempBin.push(1);
	            } else {
	              tempBin.push(0);
	            }
	          }
	        } else {
	          blackMax = [0, 0];
	          whiteMax = [0, 0];
	          for (i = 0; i < BinArray.length; i++) {
	            if (i % 2 === 0) {
	              if (BinArray[i] > blackMax[0]) {
	                blackMax[0] = BinArray[i];
	                blackMax[1] = i;
	              }
	            } else {
	              if (BinArray[i] > whiteMax[0]) {
	                whiteMax[0] = BinArray[i];
	                whiteMax[1] = i;
	              }
	            }
	          }
	          if (blackMax[0] / whiteMax[0] > 1.55) {
	            var tempArray = blackMax;
	            blackMax = [tempArray, [0, 0],
	              [0, 0]
	            ];
	            for (i = 0; i < BinArray.length; i++) {
	              if (i % 2 === 0) {
	                if (BinArray[i] > blackMax[1][0] && i !== blackMax[0][1]) {
	                  blackMax[1][0] = BinArray[i];
	                  prevPos = blackMax[1][1];
	                  blackMax[1][1] = i;
	                  i = prevPos - 1;
	                  continue;
	                }
	                if (BinArray[i] > blackMax[2][0] && i !== blackMax[0][1] && i !== blackMax[1][1]) {
	                  blackMax[2][0] = BinArray[i];
	                  blackMax[2][1] = i;
	                }
	              }
	            }
	            if (SecureCodabar) {
	              wideAvrg = blackMax[0][0] + blackMax[1][0] + blackMax[2][0];
	              wideAvrg /= 3;
	              for (i = 0; i < blackMax.length; i++) {
	                if (blackMax[i][0] / wideAvrg > upperTolerance || blackMax[i][0] / wideAvrg < lowerTolerance) return [];
	              }
	              narrowAvrg = 0;
	              for (i = 0; i < BinArray.length; i++) {
	                if (i === blackMax[0][1] || i === blackMax[1][1] || i === blackMax[2][1]) continue;
	                narrowAvrg += BinArray[i];
	              }
	              narrowAvrg /= 4;
	              for (i = 0; i < BinArray.length; i++) {
	                if (i === blackMax[0][1] || i === blackMax[1][1] || i === blackMax[2][1]) continue;
	                if (BinArray[i] / narrowAvrg > upperTolerance || BinArray[i] / narrowAvrg < lowerTolerance) return [];
	              }
	            }
	            for (i = 0; i < BinArray.length; i++) {
	              if (i === blackMax[0][1] || i === blackMax[1][1] || i === blackMax[2][1]) {
	                tempBin.push(1);
	              } else {
	                tempBin.push(0);
	              }
	            }
	          } else {
	            if (SecureCodabar) {
	              wideAvrg = blackMax[0] + whiteMax[0];
	              wideAvrg /= 2;
	              if (blackMax[0] / wideAvrg > 1.5 || blackMax[0] / wideAvrg < 0.5) return [];
	              if (whiteMax[0] / wideAvrg > 1.5 || whiteMax[0] / wideAvrg < 0.5) return [];
	              narrowAvrg = 0;
	              for (i = 0; i < BinArray.length; i++) {
	                if (i === blackMax[1] || i === whiteMax[1]) continue;
	                narrowAvrg += BinArray[i];
	              }
	              narrowAvrg /= 5;
	              for (i = 0; i < BinArray.length; i++) {
	                if (i === blackMax[1] || i === whiteMax[1]) continue;
	                if (BinArray[i] / narrowAvrg > upperTolerance || BinArray[i] / narrowAvrg < lowerTolerance) return [];
	              }
	            }
	            for (i = 0; i < BinArray.length; i++) {
	              if (i === blackMax[1] || i === whiteMax[1]) {
	                tempBin.push(1);
	              } else {
	                tempBin.push(0);
	              }
	            }
	          }
	        }
	        result.push(tempBin);
	        continue;
	      }
	      if (type === 4 || type === 5) {
	        max = [
	          [0, 0],
	          [0, 0]
	        ];
	        for (i = 0; i < BinArray.length; i++) {
	          if (!isFinite(BinArray[i])) return [];
	          if (BinArray[i] > max[0][0]) {
	            max[0][0] = BinArray[i];
	            prevPos = max[0][1];
	            max[0][1] = i;
	            i = prevPos - 1;
	          }
	          if (BinArray[i] > max[1][0] && i !== max[0][1]) {
	            max[1][0] = BinArray[i];
	            max[1][1] = i;
	          }
	        }
	        if (Secure2Of5) {
	          wideAvrg = max[0][0] + max[1][0];
	          wideAvrg /= 2;
	          if (max[0][0] / wideAvrg > 1.3 || max[0][0] / wideAvrg < 0.7) return [];
	          if (max[1][0] / wideAvrg > 1.3 || max[1][0] / wideAvrg < 0.7) return [];
	          narrowAvrg = 0;
	          for (i = 0; i < BinArray.length; i++) {
	            if (i === max[0][1] || i === max[1][1]) continue;
	            narrowAvrg += BinArray[i];
	          }
	          narrowAvrg /= 3;
	          for (i = 0; i < BinArray.length; i++) {
	            if (i === max[0][1] || i === max[1][1]) continue;
	            if (BinArray[i] / narrowAvrg > 1.3 || BinArray[i] / narrowAvrg < 0.7) return [];
	          }
	        }
	        for (i = 0; i < BinArray.length; i++) {
	          if (i === max[0][1] || i === max[1][1]) {
	            tempBin.push(1);
	            continue;
	          }
	          tempBin.push(0);
	        }
	        result.push(tempBin);
	        continue;
	      }
	      while (counter < totalBars) {
	        sum += BinArray[counter];
	        counter++;
	      }
	      if (type === 2) {
	        var indexCount = [];
	        blackMax = [
	          [0, 0],
	          [0, 0]
	        ];
	        whiteMax = [0, 0];
	        for (j = 0; j < BinArray.length; j++) {
	          if (j % 2 === 0) {
	            if (BinArray[j] > blackMax[0][0]) {
	              blackMax[0][0] = BinArray[j];
	              prevPos = blackMax[0][1];
	              blackMax[0][1] = j;
	              j = prevPos;
	            }
	            if (BinArray[j] > blackMax[1][0] && j !== blackMax[0][1]) {
	              blackMax[1][0] = BinArray[j];
	              blackMax[1][1] = j;
	            }
	          } else {
	            if (BinArray[j] > whiteMax[0]) {
	              whiteMax[0] = BinArray[j];
	              whiteMax[1] = j;
	            }
	          }
	        }
	        if (whiteMax[0] / blackMax[0][0] > 1.5 && whiteMax[0] / blackMax[1][0] > 1.5) {
	          blackMax = [
	            [0, 0],
	            [0, 0]
	          ];
	          for (j = 0; j < BinArray.length; j++) {
	            if (j % 2 !== 0) {
	              if (BinArray[j] > blackMax[0][0] && j !== whiteMax[1]) {
	                blackMax[0][0] = BinArray[j];
	                prevPos = blackMax[0][1];
	                blackMax[0][1] = j;
	                j = prevPos;
	              }
	              if (BinArray[j] > blackMax[1][0] && j !== blackMax[0][1] && j !== whiteMax[1]) {
	                blackMax[1][0] = BinArray[j];
	                blackMax[1][1] = j;
	              }
	            }
	          }
	        }
	        wideAvrg = blackMax[0][0] + blackMax[1][0] + whiteMax[0];
	        wideAvrg /= 3;
	        if (blackMax[0][0] / wideAvrg > 1.6 || blackMax[0][0] / wideAvrg < 0.4) return [];
	        if (blackMax[1][0] / wideAvrg > 1.6 || blackMax[1][0] / wideAvrg < 0.4) return [];
	        if (whiteMax[0] / wideAvrg > 1.6 || whiteMax[0] / wideAvrg < 0.4) return [];
	        narrowAvrg = 0;
	        for (i = 0; i < BinArray.length; i++) {
	          if (i === blackMax[0][1] || i === blackMax[1][1] || i === whiteMax[1]) continue;
	          narrowAvrg += BinArray[i];
	        }
	        narrowAvrg /= 6;
	        for (i = 0; i < BinArray.length; i++) {
	          if (i === blackMax[0][1] || i === blackMax[1][1] || i === whiteMax[1]) continue;
	          if (BinArray[i] / narrowAvrg > 1.6 || BinArray[i] / narrowAvrg < 0.4) return [];
	        }
	        for (j = 0; j < BinArray.length; j++) {
	          if (j === blackMax[0][1] || j === blackMax[1][1] || j === whiteMax[1]) {
	            tempBin.push(2);
	          } else {
	            tempBin.push(1);
	          }
	        }
	        result.push(tempBin);
	        continue;
	      }
	      if (type === 3) {
	        max = [
	          [0, 0],
	          [0, 0],
	          [0, 0]
	        ];
	        for (j = 0; j < BinArray.length; j++) {
	          if (BinArray[j] > max[0][0]) {
	            max[0][0] = BinArray[j];
	            prevPos = max[0][1];
	            max[0][1] = j;
	            j = prevPos;
	          }
	          if (BinArray[j] > max[1][0] && j !== max[0][1]) {
	            max[1][0] = BinArray[j];
	            prevPos = max[1][1];
	            max[1][1] = j;
	            j = prevPos;
	          }
	          if (BinArray[j] > max[2][0] && j !== max[0][1] && j !== max[1][1]) {
	            max[2][0] = BinArray[j];
	            max[2][1] = j;
	          }
	        }
	        if (max[0][0] / max[1][0] >= 3) {
	          narrowAvrg = 0;
	          for (j = 0; j < BinArray.length; j++) {
	            if (j === max[0][1]) continue;
	            narrowAvrg += BinArray[j];
	          }
	          narrowAvrg /= 3;
	          for (j = 0; j < BinArray.length; j++) {
	            if (j === max[0][1]) continue;
	            if (BinArray[j] / narrowAvrg < 0.02 || BinArray[j] / narrowAvrg > 3) return {
	              data: [],
	              correction: 0
	            };
	          }
	          if (max[0][0] / narrowAvrg < 2.2 || max[0][0] / narrowAvrg > 6) return {
	            data: [],
	            correction: 0
	          };
	          for (j = 0; j < BinArray.length; j++) {
	            if (j === max[0][1]) {
	              tempBin.push(4);
	            } else {
	              tempBin.push(1);
	            }
	          }
	          result.push(tempBin);
	        } else if (max[0][0] / max[2][0] > 2) {
	          wideAvrg = max[0][0] + max[1][0];
	          wideAvrg /= 5;
	          if (max[0][0] / (wideAvrg * 3) < 0.02 || max[0][0] / (wideAvrg * 3) > 3) return {
	            data: [],
	            correction: 0
	          };
	          if (max[1][0] / (wideAvrg * 2) < 0.02 || max[1][0] / (wideAvrg * 2) > 3) return {
	            data: [],
	            correction: 0
	          };
	          narrowAvrg = 0;
	          for (j = 0; j < BinArray.length; j++) {
	            if (j === max[0][1] || j === max[1][1]) continue;
	            narrowAvrg += BinArray[j];
	          }
	          narrowAvrg /= 2;
	          for (j = 0; j < BinArray.length; j++) {
	            if (j === max[0][1] || j === max[1][1]) continue;
	            if (BinArray[j] / narrowAvrg < 0.02 || BinArray[j] / narrowAvrg > 3) return {
	              data: [],
	              correction: 0
	            };
	          }
	          for (j = 0; j < BinArray.length; j++) {
	            if (j === max[0][1]) {
	              tempBin.push(3);
	            } else if (j === max[1][1]) {
	              tempBin.push(2);
	            } else {
	              tempBin.push(1);
	            }
	          }
	          result.push(tempBin);
	        } else {
	          if (max[0][1] % 2 === max[1][1] % 2 && max[0][1] % 2 === max[2][1] % 2) {
	            var modMem = max[0][1] % 2;
	            max[2] = [0, 0];
	            for (j = 0; j < BinArray.length; j++) {
	              if (j % 2 === modMem) continue;
	              if (BinArray[j] > max[2][0]) {
	                max[2][0] = BinArray[j];
	                max[2][1] = j;
	              }
	            }
	          }
	          wideAvrg = max[0][0] + max[1][0] + max[2][0];
	          wideAvrg /= 3;
	          for (j = 0; j < max.length; j++) {
	            if (max[j][0] / wideAvrg < 0.02 || max[j][0] / wideAvrg > 3) return {
	              data: [],
	              correction: 0
	            };
	          }
	          var narrow = 0;
	          for (j = 0; j < BinArray.length; j++) {
	            if (j === max[0][1] || j === max[1][1] || j === max[2][1]) continue;
	            narrow = BinArray[j];
	          }
	          if (wideAvrg / narrow < 0.02 || wideAvrg / narrow > 3) return {
	            data: [],
	            correction: 0
	          };
	          for (j = 0; j < BinArray.length; j++) {
	            if (j === max[0][1] || j === max[1][1] || j === max[2][1]) {
	              tempBin.push(2);
	            } else {
	              tempBin.push(1);
	            }
	          }
	          result.push(tempBin);
	        }
	        for (j = 0; j < tempBin.length; j++) {
	          testData += Math.abs(tempBin[j] - (BinArray[j] / sum) * total);
	        }
	        continue;
	      }
	      counter = 0;
	      while (counter < totalBars) {
	        tempBin.push((BinArray[counter] / sum) * total);
	        counter++;
	      }
	      counter = 0;
	      while (counter < totalBars) {
	        tempBin[counter] = tempBin[counter] > maxLength ? maxLength : tempBin[counter];
	        tempBin[counter] = tempBin[counter] < 1 ? 1 : tempBin[counter];
	        tempBin[counter] = Math.round(tempBin[counter]);
	        counter++;
	      }
	      if (type === 3) {
	        var checking = 0;
	        for (i = 0; i < tempBin.length; i++) {
	          checking += tempBin[i];
	        }
	        if (checking > 7) {
	          max = 0;
	          var hitIndex = 0;
	          for (i = 0; i < tempBin.length; i++) {
	            if (tempBin[i] > max) {
	              max = tempBin[i];
	              hitIndex = i;
	            }
	          }
	          tempBin[hitIndex] = max - (checking - 7);
	        }
	      }
	      if (type === 3) {
	        for (i = 0; i < tempBin.length; i++) {
	          testData += Math.abs(tempBin[i] - (BinArray[i] / sum) * total);
	        }
	      }
	      result.push(tempBin);
	    }
	    if (type === 3) {
	      return {
	        data: result,
	        correction: testData
	      };
	    } else {
	      return result;
	    }
	  }

	  function CheckCode128(string) {
	    var checksum = string[string.length - 2].join("");
	    var i;
	    checksum = Code128Encoding.value.indexOf(checksum);
	    if (checksum === -1) return false;
	    var summarizer = Code128Encoding.value.indexOf(string[0].join(""));
	    if (summarizer === -1) return false;
	    var startChar = Code128Encoding[string[0].join("")];
	    if (typeof startChar === 'undefined') return false;
	    if (startChar !== "A" && startChar !== "B" && startChar !== "C") return false;
	    for (i = 1; i < (string.length - 2); i++) {
	      summarizer += Code128Encoding.value.indexOf(string[i].join("")) * i;
	      if (Code128Encoding.value.indexOf(string[i].join("")) === -1) return false;
	    }
	    return (summarizer % 103 === checksum);
	  }

	  function Decode2Of5(string) {
	    var result = "";
	    var i;
	    for (i = 0; i < string.length; i++) {
	      if (TwoOfFiveEncoding.indexOf(string[i].join("")) === -1) return false;
	      result += TwoOfFiveEncoding.indexOf(string[i].join(""));
	    }
	    return result;
	  }

	  function DecodeCodaBar(string) {
	    var result = "";
	    var start = string[0].join("");
	    var end = string[string.length - 1].join("");
	    var i;
	    if (!(CodaBarEncoding[start] === "A" || CodaBarEncoding[start] === "B" || CodaBarEncoding[start] === "C" || CodaBarEncoding[start] === "D")) return false;
	    if (!(CodaBarEncoding[end] === "A" || CodaBarEncoding[end] === "B" || CodaBarEncoding[end] === "C" || CodaBarEncoding[end] === "D")) return false;
	    for (i = 1; i < string.length - 1; i++) {
	      if (typeof CodaBarEncoding[string[i].join("")] === 'undefined') return false;
	      result += CodaBarEncoding[string[i].join("")];
	    }
	    return result;
	  }

	  function DecodeEAN13(string) {
	    if (string.length !== 12) return false;
	    var leftSide = string.slice(0, 6);
	    var trigger = false;
	    var rightSide = string.slice(6, string.length);
	    var i;
	    for (i = 0; i < leftSide.length; i++) {
	      leftSide[i] = leftSide[i].join("");
	      if (leftSide[i].length !== 4) {
	        trigger = true;
	        break;
	      }
	    }
	    if (trigger) return false;
	    for (i = 0; i < rightSide.length; i++) {
	      rightSide[i] = rightSide[i].join("");
	      if (rightSide[i].length !== 4) {
	        trigger = true;
	        break;
	      }
	    }
	    if (trigger) return false;
	    var decodeFormat = [];
	    for (i = 0; i < leftSide.length; i++) {
	      if (typeof EAN13Encoding.L[leftSide[i]] !== 'undefined') {
	        decodeFormat.push("L");
	      } else if (typeof EAN13Encoding.G[leftSide[i]] !== 'undefined') {
	        decodeFormat.push("G");
	      } else {
	        trigger = true;
	        break;
	      }
	    }
	    if (trigger) return false;
	    var resultArray = [];
	    if (typeof EAN13Encoding.formats[decodeFormat.join("")] === 'undefined') return false;
	    resultArray.push(EAN13Encoding.formats[decodeFormat.join("")]);
	    for (i = 0; i < leftSide.length; i++) {
	      if (typeof EAN13Encoding[decodeFormat[i]][leftSide[i]] === 'undefined') {
	        trigger = true;
	        break;
	      }
	      resultArray.push(EAN13Encoding[decodeFormat[i]][leftSide[i]]);
	    }
	    if (trigger) return false;
	    for (i = 0; i < rightSide.length; i++) {
	      if (typeof EAN13Encoding.R[rightSide[i]] === 'undefined') {
	        trigger = true;
	        break;
	      }
	      resultArray.push(EAN13Encoding.R[rightSide[i]]);
	    }
	    if (trigger) return false;
	    var weight = 3;
	    var sum = 0;
	    for (i = resultArray.length - 2; i >= 0; i--) {
	      sum += resultArray[i] * weight;
	      if (weight === 3) {
	        weight = 1;
	      } else {
	        weight = 3;
	      }
	    }
	    sum = (10 - sum % 10) % 10;
	    if (resultArray[resultArray.length - 1] === sum) {
	      return resultArray.join("");
	    } else {
	      return false;
	    }
	  }

	  function CheckCode93(string) {
	    var checkOne = string[string.length - 3].join("");
	    var checkTwo = string[string.length - 2].join("");
	    var failSafe = true;
	    if (typeof Code93Encoding[checkOne] === 'undefined') return false;
	    if (typeof Code93Encoding[checkTwo] === 'undefined') return false;
	    var checkSum = Code93Encoding[checkOne].value;
	    var weight = 1;
	    var sum = 0;
	    var i;
	    for (i = string.length - 4; i > 0; i--) {
	      failSafe = typeof Code93Encoding[string[i].join("")] === 'undefined' ? false : failSafe;
	      if (!failSafe) break;
	      sum += Code93Encoding[string[i].join("")].value * weight;
	      weight++;
	      if (weight > 20) weight = 1;
	    }
	    var firstCheck = sum % 47;
	    var firstBool = firstCheck === checkSum;
	    if (!firstBool) return false;
	    if (!failSafe) return false;
	    sum = firstCheck;
	    weight = 2;
	    checkSum = Code93Encoding[checkTwo].value;
	    for (i = string.length - 4; i > 0; i--) {
	      failSafe = typeof Code93Encoding[string[i].join("")] === 'undefined' ? false : failSafe;
	      if (!failSafe) break;
	      sum += Code93Encoding[string[i].join("")].value * weight;
	      weight++;
	      if (weight > 15) weight = 1;
	    }
	    var secondCheck = sum % 47;
	    var secondBool = secondCheck === checkSum;
	    return secondBool && firstBool;
	  }

	  function CheckCode39(string) {
	    var trigger = true;
	    if (typeof Code39Encoding[string[0].join("")] === 'undefined') return false;
	    if (Code39Encoding[string[0].join("")].character !== "*") return false;
	    if (typeof Code39Encoding[string[string.length - 1].join("")] === 'undefined') return false;
	    if (Code39Encoding[string[string.length - 1].join("")].character !== "*") return false;
	    for (i = 1; i < string.length - 1; i++) {
	      if (typeof Code39Encoding[string[i].join("")] === 'undefined') {
	        trigger = false;
	        break;
	      }
	    }
	    return trigger;
	  }

	  function DecodeCode39(string) {
	    var resultString = "";
	    var special = false;
	    var character = "";
	    var specialchar = "";
	    for (i = 1; i < string.length - 1; i++) {
	      character = Code39Encoding[string[i].join("")].character;
	      if (character === "$" || character === "/" || character === "+" || character === "%") {
	        // if next character exists => this a special character
	        if (i + 1 < string.length - 1) {
	          special = true;
	          specialchar = character;
	          continue;
	        }
	      }
	      if (special) {
	        if (typeof ExtendedEncoding[specialchar + character] === 'undefined') {} else {
	          resultString += ExtendedEncoding[specialchar + character];
	        }
	        special = false;
	        continue;
	      }
	      resultString += character;
	    }
	    return resultString;
	  }

	  function DecodeCode93(string) {
	    var resultString = "";
	    var special = false;
	    var character = "";
	    var specialchar = "";
	    for (i = 1; i < string.length - 3; i++) {
	      character = Code93Encoding[string[i].join("")].character;
	      if (character === "($)" || character === "(/)" || character === "(+)" || character === "(%)") {
	        special = true;
	        specialchar = character[1];
	        continue;
	      }
	      if (special) {
	        if (typeof ExtendedEncoding[specialchar + character] === 'undefined') {} else {
	          resultString += ExtendedEncoding[specialchar + character];
	        }
	        special = false;
	        continue;
	      }
	      resultString += character;
	    }
	    return resultString;
	  }

	  function DecodeCode128(string) {
	    var set = Code128Encoding[string[0].join("")];
	    var symbol;
	    var Code128Format = "Code128";
	    var resultString = "";
	    var i;
	    for (i = 1; i < (string.length - 2); i++) {
	      symbol = Code128Encoding[string[i].join("")][set];
	      switch (symbol) {
	        case "FNC1":
	          if (i === 1) Code128Format = "GS1-128";
	          break;
	        case "FNC2":
	        case "FNC3":
	        case "FNC4":
	          break;
	        case "SHIFT_B":
	          i++;
	          resultString += Code128Encoding[string[i].join("")].B;
	          break;
	        case "SHIFT_A":
	          i++;
	          resultString += Code128Encoding[string[i].join("")].A;
	          break;
	        case "Code_A":
	          set = "A";
	          break;
	        case "Code_B":
	          set = "B";
	          break;
	        case "Code_C":
	          set = "C";
	          break;
	        default:
	          resultString += symbol;
	      }
	    }
	    return {
	      string: resultString,
	      format: Code128Format
	    };
	  }
	  TwoOfFiveEncoding = ["00110", "10001", "01001", "11000", "00101", "10100", "01100", "00011", "10010", "01010"];
	  Code128Encoding = {
	    "212222": {
	      A: " ",
	      B: " ",
	      C: "00"
	    },
	    "222122": {
	      A: "!",
	      B: "!",
	      C: "01"
	    },
	    "222221": {
	      A: '"',
	      B: '"',
	      C: "02"
	    },
	    "121223": {
	      A: "#",
	      B: "#",
	      C: "03"
	    },
	    "121322": {
	      A: "$",
	      B: "$",
	      C: "04"
	    },
	    "131222": {
	      A: "%",
	      B: "%",
	      C: "05"
	    },
	    "122213": {
	      A: "&",
	      B: "&",
	      C: "06"
	    },
	    "122312": {
	      A: "'",
	      B: "'",
	      C: "07"
	    },
	    "132212": {
	      A: "(",
	      B: "(",
	      C: "08"
	    },
	    "221213": {
	      A: ")",
	      B: ")",
	      C: "09"
	    },
	    "221312": {
	      A: "*",
	      B: "*",
	      C: "10"
	    },
	    "231212": {
	      A: "+",
	      B: "+",
	      C: "11"
	    },
	    "112232": {
	      A: ",",
	      B: ",",
	      C: "12"
	    },
	    "122132": {
	      A: "-",
	      B: "-",
	      C: "13"
	    },
	    "122231": {
	      A: ".",
	      B: ".",
	      C: "14"
	    },
	    "113222": {
	      A: "/",
	      B: "/",
	      C: "15"
	    },
	    "123122": {
	      A: "0",
	      B: "0",
	      C: "16"
	    },
	    "123221": {
	      A: "1",
	      B: "1",
	      C: "17"
	    },
	    "223211": {
	      A: "2",
	      B: "2",
	      C: "18"
	    },
	    "221132": {
	      A: "3",
	      B: "3",
	      C: "19"
	    },
	    "221231": {
	      A: "4",
	      B: "4",
	      C: "20"
	    },
	    "213212": {
	      A: "5",
	      B: "5",
	      C: "21"
	    },
	    "223112": {
	      A: "6",
	      B: "6",
	      C: "22"
	    },
	    "312131": {
	      A: "7",
	      B: "7",
	      C: "23"
	    },
	    "311222": {
	      A: "8",
	      B: "8",
	      C: "24"
	    },
	    "321122": {
	      A: "9",
	      B: "9",
	      C: "25"
	    },
	    "321221": {
	      A: ":",
	      B: ":",
	      C: "26"
	    },
	    "312212": {
	      A: ";",
	      B: ";",
	      C: "27"
	    },
	    "322112": {
	      A: "<",
	      B: "<",
	      C: "28"
	    },
	    "322211": {
	      A: "=",
	      B: "=",
	      C: "29"
	    },
	    "212123": {
	      A: ">",
	      B: ">",
	      C: "30"
	    },
	    "212321": {
	      A: "?",
	      B: "?",
	      C: "31"
	    },
	    "232121": {
	      A: "@",
	      B: "@",
	      C: "32"
	    },
	    "111323": {
	      A: "A",
	      B: "A",
	      C: "33"
	    },
	    "131123": {
	      A: "B",
	      B: "B",
	      C: "34"
	    },
	    "131321": {
	      A: "C",
	      B: "C",
	      C: "35"
	    },
	    "112313": {
	      A: "D",
	      B: "D",
	      C: "36"
	    },
	    "132113": {
	      A: "E",
	      B: "E",
	      C: "37"
	    },
	    "132311": {
	      A: "F",
	      B: "F",
	      C: "38"
	    },
	    "211313": {
	      A: "G",
	      B: "G",
	      C: "39"
	    },
	    "231113": {
	      A: "H",
	      B: "H",
	      C: "40"
	    },
	    "231311": {
	      A: "I",
	      B: "I",
	      C: "41"
	    },
	    "112133": {
	      A: "J",
	      B: "J",
	      C: "42"
	    },
	    "112331": {
	      A: "K",
	      B: "K",
	      C: "43"
	    },
	    "132131": {
	      A: "L",
	      B: "L",
	      C: "44"
	    },
	    "113123": {
	      A: "M",
	      B: "M",
	      C: "45"
	    },
	    "113321": {
	      A: "N",
	      B: "N",
	      C: "46"
	    },
	    "133121": {
	      A: "O",
	      B: "O",
	      C: "47"
	    },
	    "313121": {
	      A: "P",
	      B: "P",
	      C: "48"
	    },
	    "211331": {
	      A: "Q",
	      B: "Q",
	      C: "49"
	    },
	    "231131": {
	      A: "R",
	      B: "R",
	      C: "50"
	    },
	    "213113": {
	      A: "S",
	      B: "S",
	      C: "51"
	    },
	    "213311": {
	      A: "T",
	      B: "T",
	      C: "52"
	    },
	    "213131": {
	      A: "U",
	      B: "U",
	      C: "53"
	    },
	    "311123": {
	      A: "V",
	      B: "V",
	      C: "54"
	    },
	    "311321": {
	      A: "W",
	      B: "W",
	      C: "55"
	    },
	    "331121": {
	      A: "X",
	      B: "X",
	      C: "56"
	    },
	    "312113": {
	      A: "Y",
	      B: "Y",
	      C: "57"
	    },
	    "312311": {
	      A: "Z",
	      B: "Z",
	      C: "58"
	    },
	    "332111": {
	      A: "[",
	      B: "[",
	      C: "59"
	    },
	    "314111": {
	      A: "\\",
	      B: "\\",
	      C: "60"
	    },
	    "221411": {
	      A: "]",
	      B: "]",
	      C: "61"
	    },
	    "431111": {
	      A: "^",
	      B: "^",
	      C: "62"
	    },
	    "111224": {
	      A: "_",
	      B: "_",
	      C: "63"
	    },
	    "111422": {
	      A: "NUL",
	      B: "`",
	      C: "64"
	    },
	    "121124": {
	      A: "SOH",
	      B: "a",
	      C: "65"
	    },
	    "121421": {
	      A: "STX",
	      B: "b",
	      C: "66"
	    },
	    "141122": {
	      A: "ETX",
	      B: "c",
	      C: "67"
	    },
	    "141221": {
	      A: "EOT",
	      B: "d",
	      C: "68"
	    },
	    "112214": {
	      A: "ENQ",
	      B: "e",
	      C: "69"
	    },
	    "112412": {
	      A: "ACK",
	      B: "f",
	      C: "70"
	    },
	    "122114": {
	      A: "BEL",
	      B: "g",
	      C: "71"
	    },
	    "122411": {
	      A: "BS",
	      B: "h",
	      C: "72"
	    },
	    "142112": {
	      A: "HT",
	      B: "i",
	      C: "73"
	    },
	    "142211": {
	      A: "LF",
	      B: "j",
	      C: "74"
	    },
	    "241211": {
	      A: "VT",
	      B: "k",
	      C: "75"
	    },
	    "221114": {
	      A: "FF",
	      B: "l",
	      C: "76"
	    },
	    "413111": {
	      A: "CR",
	      B: "m",
	      C: "77"
	    },
	    "241112": {
	      A: "SO",
	      B: "n",
	      C: "78"
	    },
	    "134111": {
	      A: "SI",
	      B: "o",
	      C: "79"
	    },
	    "111242": {
	      A: "DLE",
	      B: "p",
	      C: "80"
	    },
	    "121142": {
	      A: "DC1",
	      B: "q",
	      C: "81"
	    },
	    "121241": {
	      A: "DC2",
	      B: "r",
	      C: "82"
	    },
	    "114212": {
	      A: "DC3",
	      B: "s",
	      C: "83"
	    },
	    "124112": {
	      A: "DC4",
	      B: "t",
	      C: "84"
	    },
	    "124211": {
	      A: "NAK",
	      B: "u",
	      C: "85"
	    },
	    "411212": {
	      A: "SYN",
	      B: "v",
	      C: "86"
	    },
	    "421112": {
	      A: "ETB",
	      B: "w",
	      C: "87"
	    },
	    "421211": {
	      A: "CAN",
	      B: "x",
	      C: "88"
	    },
	    "212141": {
	      A: "EM",
	      B: "y",
	      C: "89"
	    },
	    "214121": {
	      A: "SUB",
	      B: "z",
	      C: "90"
	    },
	    "412121": {
	      A: "ESC",
	      B: "{",
	      C: "91"
	    },
	    "111143": {
	      A: "FS",
	      B: "|",
	      C: "92"
	    },
	    "111341": {
	      A: "GS",
	      B: "}",
	      C: "93"
	    },
	    "131141": {
	      A: "RS",
	      B: "~",
	      C: "94"
	    },
	    "114113": {
	      A: "US",
	      B: "DEL",
	      C: "95"
	    },
	    "114311": {
	      A: "FNC3",
	      B: "FNC3",
	      C: "96"
	    },
	    "411113": {
	      A: "FNC2",
	      B: "FNC2",
	      C: "97"
	    },
	    "411311": {
	      A: "SHIFT_B",
	      B: "SHIFT_A",
	      C: "98"
	    },
	    "113141": {
	      A: "Code_C",
	      B: "Code_C",
	      C: "99"
	    },
	    "114131": {
	      A: "Code_B",
	      B: "FNC4",
	      C: "Code_B"
	    },
	    "311141": {
	      A: "FNC4",
	      B: "Code_A",
	      C: "Code_A"
	    },
	    "411131": {
	      A: "FNC1",
	      B: "FNC1",
	      C: "FNC1"
	    },
	    "211412": "A",
	    "211214": "B",
	    "211232": "C",
	    "233111": {
	      A: "STOP",
	      B: "STOP",
	      C: "STOP"
	    },
	    value: [
	      "212222",
	      "222122",
	      "222221",
	      "121223",
	      "121322",
	      "131222",
	      "122213",
	      "122312",
	      "132212",
	      "221213",
	      "221312",
	      "231212",
	      "112232",
	      "122132",
	      "122231",
	      "113222",
	      "123122",
	      "123221",
	      "223211",
	      "221132",
	      "221231",
	      "213212",
	      "223112",
	      "312131",
	      "311222",
	      "321122",
	      "321221",
	      "312212",
	      "322112",
	      "322211",
	      "212123",
	      "212321",
	      "232121",
	      "111323",
	      "131123",
	      "131321",
	      "112313",
	      "132113",
	      "132311",
	      "211313",
	      "231113",
	      "231311",
	      "112133",
	      "112331",
	      "132131",
	      "113123",
	      "113321",
	      "133121",
	      "313121",
	      "211331",
	      "231131",
	      "213113",
	      "213311",
	      "213131",
	      "311123",
	      "311321",
	      "331121",
	      "312113",
	      "312311",
	      "332111",
	      "314111",
	      "221411",
	      "431111",
	      "111224",
	      "111422",
	      "121124",
	      "121421",
	      "141122",
	      "141221",
	      "112214",
	      "112412",
	      "122114",
	      "122411",
	      "142112",
	      "142211",
	      "241211",
	      "221114",
	      "413111",
	      "241112",
	      "134111",
	      "111242",
	      "121142",
	      "121241",
	      "114212",
	      "124112",
	      "124211",
	      "411212",
	      "421112",
	      "421211",
	      "212141",
	      "214121",
	      "412121",
	      "111143",
	      "111341",
	      "131141",
	      "114113",
	      "114311",
	      "411113",
	      "411311",
	      "113141",
	      "114131",
	      "311141",
	      "411131",
	      "211412",
	      "211214",
	      "211232",
	      "233111"
	    ]
	  };

	  Code93Encoding = {
	    "131112": {
	      value: 0,
	      character: "0"
	    },
	    "111213": {
	      value: 1,
	      character: "1"
	    },
	    "111312": {
	      value: 2,
	      character: "2"
	    },
	    "111411": {
	      value: 3,
	      character: "3"
	    },
	    "121113": {
	      value: 4,
	      character: "4"
	    },
	    "121212": {
	      value: 5,
	      character: "5"
	    },
	    "121311": {
	      value: 6,
	      character: "6"
	    },
	    "111114": {
	      value: 7,
	      character: "7"
	    },
	    "131211": {
	      value: 8,
	      character: "8"
	    },
	    "141111": {
	      value: 9,
	      character: "9"
	    },
	    "211113": {
	      value: 10,
	      character: "A"
	    },
	    "211212": {
	      value: 11,
	      character: "B"
	    },
	    "211311": {
	      value: 12,
	      character: "C"
	    },
	    "221112": {
	      value: 13,
	      character: "D"
	    },
	    "221211": {
	      value: 14,
	      character: "E"
	    },
	    "231111": {
	      value: 15,
	      character: "F"
	    },
	    "112113": {
	      value: 16,
	      character: "G"
	    },
	    "112212": {
	      value: 17,
	      character: "H"
	    },
	    "112311": {
	      value: 18,
	      character: "I"
	    },
	    "122112": {
	      value: 19,
	      character: "J"
	    },
	    "132111": {
	      value: 20,
	      character: "K"
	    },
	    "111123": {
	      value: 21,
	      character: "L"
	    },
	    "111222": {
	      value: 22,
	      character: "M"
	    },
	    "111321": {
	      value: 23,
	      character: "N"
	    },
	    "121122": {
	      value: 24,
	      character: "O"
	    },
	    "131121": {
	      value: 25,
	      character: "P"
	    },
	    "212112": {
	      value: 26,
	      character: "Q"
	    },
	    "212211": {
	      value: 27,
	      character: "R"
	    },
	    "211122": {
	      value: 28,
	      character: "S"
	    },
	    "211221": {
	      value: 29,
	      character: "T"
	    },
	    "221121": {
	      value: 30,
	      character: "U"
	    },
	    "222111": {
	      value: 31,
	      character: "V"
	    },
	    "112122": {
	      value: 32,
	      character: "W"
	    },
	    "112221": {
	      value: 33,
	      character: "X"
	    },
	    "122121": {
	      value: 34,
	      character: "Y"
	    },
	    "123111": {
	      value: 35,
	      character: "Z"
	    },
	    "121131": {
	      value: 36,
	      character: "-"
	    },
	    "311112": {
	      value: 37,
	      character: "."
	    },
	    "311211": {
	      value: 38,
	      character: " "
	    },
	    "321111": {
	      value: 39,
	      character: "$"
	    },
	    "112131": {
	      value: 40,
	      character: "/"
	    },
	    "113121": {
	      value: 41,
	      character: "+"
	    },
	    "211131": {
	      value: 42,
	      character: "%"
	    },
	    "121221": {
	      value: 43,
	      character: "($)"
	    },
	    "312111": {
	      value: 44,
	      character: "(%)"
	    },
	    "311121": {
	      value: 45,
	      character: "(/)"
	    },
	    "122211": {
	      value: 46,
	      character: "(+)"
	    },
	    "111141": {
	      value: -1,
	      character: "*"
	    }
	  };
	  Code39Encoding = {
	    "111221211": {
	      value: 0,
	      character: "0"
	    },
	    "211211112": {
	      value: 1,
	      character: "1"
	    },
	    "112211112": {
	      value: 2,
	      character: "2"
	    },
	    "212211111": {
	      value: 3,
	      character: "3"
	    },
	    "111221112": {
	      value: 4,
	      character: "4"
	    },
	    "211221111": {
	      value: 5,
	      character: "5"
	    },
	    "112221111": {
	      value: 6,
	      character: "6"
	    },
	    "111211212": {
	      value: 7,
	      character: "7"
	    },
	    "211211211": {
	      value: 8,
	      character: "8"
	    },
	    "112211211": {
	      value: 9,
	      character: "9"
	    },
	    "211112112": {
	      value: 10,
	      character: "A"
	    },
	    "112112112": {
	      value: 11,
	      character: "B"
	    },
	    "212112111": {
	      value: 12,
	      character: "C"
	    },
	    "111122112": {
	      value: 13,
	      character: "D"
	    },
	    "211122111": {
	      value: 14,
	      character: "E"
	    },
	    "112122111": {
	      value: 15,
	      character: "F"
	    },
	    "111112212": {
	      value: 16,
	      character: "G"
	    },
	    "211112211": {
	      value: 17,
	      character: "H"
	    },
	    "112112211": {
	      value: 18,
	      character: "I"
	    },
	    "111122211": {
	      value: 19,
	      character: "J"
	    },
	    "211111122": {
	      value: 20,
	      character: "K"
	    },
	    "112111122": {
	      value: 21,
	      character: "L"
	    },
	    "212111121": {
	      value: 22,
	      character: "M"
	    },
	    "111121122": {
	      value: 23,
	      character: "N"
	    },
	    "211121121": {
	      value: 24,
	      character: "O"
	    },
	    "112121121": {
	      value: 25,
	      character: "P"
	    },
	    "111111222": {
	      value: 26,
	      character: "Q"
	    },
	    "211111221": {
	      value: 27,
	      character: "R"
	    },
	    "112111221": {
	      value: 28,
	      character: "S"
	    },
	    "111121221": {
	      value: 29,
	      character: "T"
	    },
	    "221111112": {
	      value: 30,
	      character: "U"
	    },
	    "122111112": {
	      value: 31,
	      character: "V"
	    },
	    "222111111": {
	      value: 32,
	      character: "W"
	    },
	    "121121112": {
	      value: 33,
	      character: "X"
	    },
	    "221121111": {
	      value: 34,
	      character: "Y"
	    },
	    "122121111": {
	      value: 35,
	      character: "Z"
	    },
	    "121111212": {
	      value: 36,
	      character: "-"
	    },
	    "221111211": {
	      value: 37,
	      character: "."
	    },
	    "122111211": {
	      value: 38,
	      character: " "
	    },
	    "121212111": {
	      value: 39,
	      character: "$"
	    },
	    "121211121": {
	      value: 40,
	      character: "/"
	    },
	    "121112121": {
	      value: 41,
	      character: "+"
	    },
	    "111212121": {
	      value: 42,
	      character: "%"
	    },
	    "121121211": {
	      value: -1,
	      character: "*"
	    }
	  };

	  ExtendedEncoding = {
	    "/A": '!',
	    "/B": '"',
	    "/C": '#',
	    "/D": '$',
	    "/E": '%',
	    "/F": '&',
	    "/G": "'",
	    "/H": '(',
	    "/I": ')',
	    "/J": '*',
	    "/K": '+',
	    "/L": ',',
	    "/O": '/',
	    "/Z": ':',
	    "%F": ';',
	    "%G": '<',
	    "%H": '=',
	    "%I": '>',
	    "%J": '?',
	    "%K": '[',
	    "%L": "\\",
	    "%M": ']',
	    "%N": '^',
	    "%O": '_',
	    "+A": 'a',
	    "+B": 'b',
	    "+C": 'c',
	    "+D": 'd',
	    "+E": 'e',
	    "+F": 'f',
	    "+G": 'g',
	    "+H": 'h',
	    "+I": 'i',
	    "+J": 'j',
	    "+K": 'k',
	    "+L": 'l',
	    "+M": 'm',
	    "+N": 'n',
	    "+O": 'o',
	    "+P": 'p',
	    "+Q": 'q',
	    "+R": 'r',
	    "+S": 's',
	    "+T": 't',
	    "+U": 'u',
	    "+V": 'v',
	    "+W": 'w',
	    "+X": 'x',
	    "+Y": 'y',
	    "+Z": 'z',
	    "%P": "{",
	    "%Q": '|',
	    "%R": '|',
	    "%S": '~',
	  };

	  CodaBarEncoding = {
	    "0000011": "0",
	    "0000110": "1",
	    "0001001": "2",
	    "1100000": "3",
	    "0010010": "4",
	    "1000010": "5",
	    "0100001": "6",
	    "0100100": "7",
	    "0110000": "8",
	    "1001000": "9",
	    "0001100": "-",
	    "0011000": "$",
	    "1000101": ":",
	    "1010001": "/",
	    "1010100": ".",
	    "0011111": "+",
	    "0011010": "A",
	    "0001011": "B",
	    "0101001": "C",
	    "0001110": "D"
	  };

	  EAN13Encoding = {
	    "L": {
	      "3211": 0,
	      "2221": 1,
	      "2122": 2,
	      "1411": 3,
	      "1132": 4,
	      "1231": 5,
	      "1114": 6,
	      "1312": 7,
	      "1213": 8,
	      "3112": 9
	    },
	    "G": {
	      "1123": 0,
	      "1222": 1,
	      "2212": 2,
	      "1141": 3,
	      "2311": 4,
	      "1321": 5,
	      "4111": 6,
	      "2131": 7,
	      "3121": 8,
	      "2113": 9
	    },
	    "R": {
	      "3211": 0,
	      "2221": 1,
	      "2122": 2,
	      "1411": 3,
	      "1132": 4,
	      "1231": 5,
	      "1114": 6,
	      "1312": 7,
	      "1213": 8,
	      "3112": 9
	    },
	    formats: {
	      "LLLLLL": 0,
	      "LLGLGG": 1,
	      "LLGGLG": 2,
	      "LLGGGL": 3,
	      "LGLLGG": 4,
	      "LGGLLG": 5,
	      "LGGGLL": 6,
	      "LGLGLG": 7,
	      "LGLGGL": 8,
	      "LGGLGL": 9
	    }
	  };

	  self.onmessage = function(e) {
	    var width;
	    var i;

	    ScanImage = {
	      data: new Uint8ClampedArray(e.data.scan),
	      width: e.data.scanWidth,
	      height: e.data.scanHeight
	    };
	    switch (e.data.rotation) {
	      case 8:
	        ScanImage.data = Rotate(ScanImage.data, ScanImage.width, ScanImage.height, -90);
	        width = e.data.scanWidth;
	        ScanImage.width = ScanImage.height;
	        ScanImage.height = width;
	        break;
	      case 6:
	        ScanImage.data = Rotate(ScanImage.data, ScanImage.width, ScanImage.height, 90);
	        width = e.data.scanWidth;
	        ScanImage.width = ScanImage.height;
	        ScanImage.height = width;
	        break;
	      case 3:
	        ScanImage.data = Rotate(ScanImage.data, ScanImage.width, ScanImage.height, 180);
	    }
	    Image = {
	      data: Scale(ScanImage.data, ScanImage.width, ScanImage.height),
	      width: ScanImage.width / 2,
	      height: ScanImage.height / 2
	    };
	    if (e.data.postOrientation) {
	      postMessage({
	        result: Image,
	        success: "orientationData"
	      });
	    }
	    availableFormats = ["Code128", "Code93", "Code39", "EAN-13", "2Of5", "Inter2Of5", "Codabar"];
	    FormatPriority = [];
	    var decodeFormats = ["Code128", "Code93", "Code39", "EAN-13", "2Of5", "Inter2Of5", "Codabar"];
	    SecureCodabar = true;
	    Secure2Of5 = true;
	    Multiple = true;
	    if (typeof e.data.multiple !== 'undefined') {
	      Multiple = e.data.multiple;
	    }
	    if (typeof e.data.decodeFormats !== 'undefined') {
	      decodeFormats = e.data.decodeFormats;
	    }
	    for (i = 0; i < decodeFormats.length; i++) {
	      FormatPriority.push(decodeFormats[i]);
	    }
	    CreateTable();
	    CreateScanTable();
	    var FinalResult = Main();
	    if (FinalResult.length > 0) {
	      postMessage({
	        result: FinalResult,
	        success: true
	      });
	    } else {
	      postMessage({
	        result: FinalResult,
	        success: false
	      });
	    }
	  };
	};

	var decoderWorkerBlobString = decoderWorkerBlob.toString();
	decoderWorkerBlobString = decoderWorkerBlobString.substring(decoderWorkerBlobString.indexOf("{")+1, decoderWorkerBlobString.lastIndexOf("}"));

	if (true) {
	  if (typeof module !== 'undefined' && module.exports) {
	      exports = module.exports = decoderWorkerBlobString;
	  }
	  exports.decoderWorkerBlobString = decoderWorkerBlobString;
	} else {
	  root.decoderWorkerBlobString = decoderWorkerBlobString;
	}

/***/ }
/******/ ]);