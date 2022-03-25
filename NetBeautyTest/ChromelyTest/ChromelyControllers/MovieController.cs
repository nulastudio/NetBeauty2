// Copyright © 2017 Chromely Projects. All rights reserved.
// Use of this source code is governed by MIT license that can be found in the LICENSE file.

#pragma warning disable CA1822

namespace ChromelyTest.ChromelyControllers;

/// <summary>
/// The movie controller.
/// </summary>
[ChromelyController(Name = "MovieController")]
public class MovieController : ChromelyController
{
    private readonly IChromelyConfiguration _config;

    /// <summary>
    /// Initializes a new instance of the <see cref="MovieController"/> class.
    /// </summary>
    public MovieController(IChromelyConfiguration config)
    {
        _config = config;
    }


    [ChromelyRoute(Path = "/democontroller/showdevtools")]
    public void ShowDevTools()
    {
        if (_config != null && !string.IsNullOrWhiteSpace(_config.DevToolsUrl))
        {
            BrowserLauncher.Open(_config.Platform, _config.DevToolsUrl);
        }
    }

    [ChromelyRoute(Path = "/democontroller/movies/get")]
    public List<MovieInfo> GetMovies()
    {
        var movieInfos = new List<MovieInfo>();
        var assemblyName = typeof(MovieInfo).Assembly.GetName().Name ?? string.Empty;

        movieInfos.Add(new MovieInfo(id: 1, title: "The Shawshank Redemption", year: 1994, votes: 678790, rating: 9.2, assembly: assemblyName));
        movieInfos.Add(new MovieInfo(id: 2, title: "The Godfather", year: 1972, votes: 511495, rating: 9.2, assembly: assemblyName));
        movieInfos.Add(new MovieInfo(id: 3, title: "The Godfather: Part II", year: 1974, votes: 319352, rating: 9.0, assembly: assemblyName));
        movieInfos.Add(new MovieInfo(id: 4, title: "The Good, the Bad and the Ugly", year: 1966, votes: 213030, rating: 8.9, assembly: assemblyName));
        movieInfos.Add(new MovieInfo(id: 5, title: "My Fair Lady", year: 1964, votes: 533848, rating: 8.9, assembly: assemblyName));
        movieInfos.Add(new MovieInfo(id: 6, title: "12 Angry Men", year: 1957, votes: 164558, rating: 8.9, assembly: assemblyName));

        return movieInfos;
    }

    [ChromelyRoute(Path = "/democontroller/movies/post")]
    public string SaveMovies(List<MovieInfo> movies)
    {
        var rowsReceived = movies != null ? movies.Count : 0;
        return $"{DateTime.Now}: {rowsReceived} rows of data successfully saved.";
    }
}



