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
        			<a href="/videos/{{.EscapedId}}">
						<img src="/thumbnails/{{.EscapedId}}.jpg">
						<div class="video-title">{{.Title}}</div>
					</a>
      			</li>
      		{{else}}
      			<li>No videos uploaded yet.</li>
      		{{end}}
    	</ul>
	</div>
	<script src="/static/script.js"></script>
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
	<div class="video-upload">
		<form class="video-upload" action="/upload" method="post" enctype="multipart/form-data">
			<label>
				Upload an mp4 video:
				<input type="file" name="file" accept="video/mp4" required />
			</label>
			<label>
				Upload a thumbnail:
				<input type="file" name="thumbnail" accept="image/jpg" />
			</label>
			<label>
				Enter a title:
				<input type="text" name="title" required/>
			</label>
			<label>
				Enter a description:
				<textarea rows="4"></textarea>
			</label>
			<div class="upload-buttons">
				<input type="submit" id="submit-button" value="Upload" />
				<input type="reset" value="Cancel">
			</div>
			<div class="uploading">
				<p id="upload-text">Uploading Video... Please Wait...</p>
				<div class="loader"></div>
			</div>
		</form>
	</div>
	<script src="/static/script.js"></script>
  </body>
</html>
`

const videoHTML = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>{{.Title}} - RonerTube</title>
	<link rel="stylesheet" href="/static/style.css">
    <script src="https://cdn.dashjs.org/latest/dash.all.min.js"></script>
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
    <h1>{{.Title}}</h1>
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
