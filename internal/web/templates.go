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
    <nav class="navbar">
		<div class="logo">RonerTube</div>
  			<ul class="nav-links">
			<li><a href="/">Home</a></li>
			<li><a href="/search">Search</a></li>
    		<li><a href="/upload-page">Upload</a></li>
    		<li><a href="/settings">Settings</a></li>
  		</ul>
  		<button class="nav-toggle" aria-label="toggle navigation">
    		<span class="hamburger"></span>
  		</button>
	</nav>
	<div class="videos-home">
    	<ul class="video-links">
      		{{range .}}
      			<li>
        			<a href="/videos/{{.EscapedId}}">{{.Id}} ({{.UploadTime}})</a>
      			</li>
      		{{else}}
      			<li>No videos uploaded yet.</li>
      		{{end}}
    	</ul>
	</div>
  </body>
</html>
`

const uploadpageHTML = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>RonerTube</title>
	<link rel="stylesheet" href="/static/style.css">
  </head>
  <body>
    <nav class="navbar">
		<div class="logo">RonerTube</div>
  			<ul class="nav-links">
			<li><a href="/">Home</a></li>
			<li><a href="/search">Search</a></li>
    		<li><a href="/upload-page">Upload</a></li>
    		<li><a href="/settings">Settings</a></li>
  		</ul>
  		<button class="nav-toggle" aria-label="toggle navigation">
    		<span class="hamburger"></span>
  		</button>
	</nav>
    <form action="/upload" method="post" enctype="multipart/form-data">
      <input type="file" name="file" accept="video/mp4" required />
      <input type="submit" value="Upload" />
    </form>
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
