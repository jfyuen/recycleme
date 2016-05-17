var BarcodeReader = require("./BarcodeReader")

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

/*BarcodeReader.SetErrorCallback(function() {
    App.stopLoading();
    $("#recycle_form").addClass("has-error");
    $("#help").text("cannot read image");
});*/

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
