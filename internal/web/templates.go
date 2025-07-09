package web

const indexHTML = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>RonerTube</title>
	<link rel="stylesheet" href="/static/style.css">
  </head>
  <body>
    <h1>Welcome to RonerTube</h1>
    <h2>Upload an MP4 Video</h2>
    <form action="/upload" method="post" enctype="multipart/form-data">
      <input type="file" name="file" accept="video/mp4" required />
      <input type="submit" value="Upload" />
    </form>
    <h2>Watchlist</h2>
    <ul>
      {{range .}}
      <li>
        <a href="/videos/{{.EscapedId}}">{{.Id}} ({{.UploadTime}})</a>
      </li>
      {{else}}
      <li>No videos uploaded yet.</li>
      {{end}}
    </ul>
  </body>
</html>
`

const videoHTML = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>{{.Id}} - RonerTube</title>
    <script src="https://cdn.dashjs.org/latest/dash.all.min.js"></script>
  </head>
  <body>
    <h1>{{.Id}}</h1>
	  <p>Uploaded at: {{.UploadedAt}}</p>

    <video id="dashPlayer" controls style="width: 640px; height: 360px"></video>
    <script>
      var url = "/content/{{.Id}}/manifest.mpd";
      var player = dashjs.MediaPlayer().create();
      player.initialize(document.querySelector("#dashPlayer"), url, false);
    </script>

    <p><a href="/">Back to Home</a></p>
  </body>
</html>
`
