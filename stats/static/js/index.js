// SYSTEM AND GILES STATS

function fetchSystemStats() {
  $.getJSON("/api/systemstats",
    function(data) {
      $('[data-id="uptime"]').html(data.uptime_sec.toHHMMSS());
      $('[data-id="cpu"]').html(data.cpu_percents + "%");
      $('[data-id="disk"]').html(data.disk_free_gb.toFixed(1)
        + " GB of " + data.disk_total_gb.toFixed(1) + " GB free");
      $('[data-id="inodes"]').html(data.inodes_free.withCommas()
        + " of " + data.inodes_total.withCommas() + " free");
      $('[data-id="memory"]').html(data.mem_free_gb.toFixed(2) + " GB of "
        + data.mem_total_gb.toFixed(2) + " GB free");
      $('[data-id="netio"]').html(data.net_sent_mb.toFixed(0) + " MB sent "
        + data.net_recv_mb.toFixed(0) + " MB received");
    })
}

function fetchGilesStats() {
  $.getJSON("/api/gilesstats",
    function(data) {
      // Iterate through data object
      for (var key in data) {
        if (data.hasOwnProperty(key)) {
          $('[data-id="' + key + '"]').html(data[key]);
        }
      }
    });
}

fetchSystemStats()
fetchGilesStats()
setInterval(function() {
  fetchSystemStats();
  fetchGilesStats();

  //TODO: Change?
  //update_timeseries_data()
  //update_timeseries_table()
}, 1000)

// TIMESERIES
// TODO: Do not hardcode api url
var API_URL = "http://localhost:8079/api/";
var querystr = "select data before now limit 5 where uuid=";
var uuid_to_data = {};
var uuids = [];
var timer;

function start_timeseries_table_timer() {
  clearTimeout(timer);
  timer = setTimeout(function() {
    update_timeseries_table();
  }, 1000)
}

function update_timeseries_data() {
  $.post(API_URL + "query", "select *;", function(data) {
    //console.log(data)
    for (var i=0; i < data.length; i++) {
      uuid = data[i].uuid;
      uuid_to_data[uuid] = data[i];
      uuids.push(uuid);
      $.post(API_URL + "query", querystr + "'" + uuid + "';", function(uuid_data) {
        uuid = uuid_data[0].uuid;
        uuid_obj = uuid_to_data[uuid];
        uuid_obj.time_scale = 1/1000;  // Currently assuming all timestamps are in ms.
        last_updated = uuid_data[0].Readings[0][0];
        uuid_obj.last_updated = last_updated;
        if (uuid_data[0].Readings.length == 5) {
          uuid_obj.data_period = (last_updated - uuid_data[0].Readings[4][0])/4 * uuid_obj.time_scale;
        }
        start_timeseries_table_timer();
      });
    }
  });
}
update_timeseries_data()

function update_timeseries_table() {
  var html = "<tr><th>Path</th><th>UUID</th><th>Average reading period</th><th>Last updated</th></tr>"
  for (var i=0; i < uuids.length; i++) {
    var uuid = uuids[i];
    var uuid_obj = uuid_to_data[uuid];
    var updated_sec_ago = Math.floor(
      ((new Date).getTime()/1000 -
      uuid_obj.last_updated * uuid_obj.time_scale));
    if (updated_sec_ago < 0) {
      console.log("WARNING: Negative updated_sec_ago", uuid_to_data[uuid].last_updated, updated_sec_ago, uuid_to_data[uuid].Properties.UnitofTime);
    }
    active_status = "success";
    if (updated_sec_ago > 60*60) {
      active_status = "danger";
    } else if (updated_sec_ago > 20*60) {
      active_status = "warning";
    }
    html += "<tr><td>" + uuid_to_data[uuid].Path + "</td>";
    html += "<td>" + uuid + "</td>";
    html += "<td>" + uuid_to_data[uuid].data_period + "s</td>";
    html += "<td class='" + active_status + "'>";
    html += updated_sec_ago.toHHMMSS() + "ago</td></tr>";
  }
  $('.ts-table').html(html);
}


// MISC

Number.prototype.toHHMMSS = function () {
    var sec_num = this;
    var hours   = Math.floor(sec_num / 3600);
    var minutes = Math.floor((sec_num - (hours * 3600)) / 60);
    var seconds = sec_num - (hours * 3600) - (minutes * 60);

    if (hours   < 10 && hours >= 0) {hours   = "0"+hours};
    if (minutes < 10) {minutes = "0"+minutes};
    if (seconds < 10) {seconds = "0"+seconds};
    var time    = hours+'h '+minutes+'m '+seconds+'s ';
    return time;
}

Number.prototype.withCommas = function () {
    return this.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}
