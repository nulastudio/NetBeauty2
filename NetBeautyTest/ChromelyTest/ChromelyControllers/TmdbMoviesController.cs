// Copyright © 2017 Chromely Projects. All rights reserved.
// Use of this source code is governed by MIT license that can be found in the LICENSE file.

#pragma warning disable CA1822
#pragma warning disable IDE1006

using System.Text.Json;

namespace ChromelyTest.ChromelyControllers;

[ChromelyController(Name = "TmdbMoviesController")]
public class TmdbMoviesController : ChromelyController
{
    private const string TmdbBaseUrl = "https://api.themoviedb.org/3/";

    private const string ChromelyTmdbApiKey = "4f457e870e91b76e02292d52a46fc445";

    private static string TmdbPopularUrl(string apiKey = ChromelyTmdbApiKey) => $"movie/popular?api_key={apiKey}&language=en-US&page=1";
    private static string TmdbTopRatedUrl(string apiKey = ChromelyTmdbApiKey) => $"movie/top_rated?api_key={apiKey}&language=en-US&page=1";
    private static string TmdbNowPlayingUrl(string apiKey = ChromelyTmdbApiKey) => $"movie/now_playing?api_key={apiKey}&language=en-US&page=1";
    private static string TmdbUpcomingUrl(string apiKey = ChromelyTmdbApiKey) => $"movie/upcoming?api_key={apiKey}&language=en-US&page=1";
    private static string TmdbSearchUrl(string queryValue, string apiKey = ChromelyTmdbApiKey) => $"search/movie?api_key={apiKey}&query={queryValue}&language=en-US&page=1&include_adult=false";
    private static string TmdbGetMovieUrl(string movieId, string apiKey = ChromelyTmdbApiKey) => $"movie/{movieId}?api_key={apiKey}";

    private readonly IChromelyConfiguration _config;

    public TmdbMoviesController(IChromelyConfiguration config)
    {
        _config = config;
    }

    [ChromelyRoute(Path = "/tmdbmoviescontroller/movies")]
    public List<Result> GetMovies(string name, string query)
    {
        if (string.IsNullOrWhiteSpace(name))
        {
            return new List<Result>();
        }

        if (name.Equals("search") && string.IsNullOrWhiteSpace(query))
        {
            return new List<Result>();
        }

        var paramUrl = string.Empty;
        switch (name.ToLower())
        {
            case "popular":
                paramUrl = TmdbPopularUrl();
                break;
            case "toprated":
                paramUrl = TmdbTopRatedUrl();
                break;
            case "nowplaying":
                paramUrl = TmdbNowPlayingUrl();
                break;
            case "upcoming":
                paramUrl = TmdbUpcomingUrl();
                break;
            case "search":
                paramUrl = TmdbSearchUrl(query);
                break;
        }

        var tmdbMoviesTask = Task.Run(() =>
        {
            return GetTmdbMovieListAsync(paramUrl);
        });

        tmdbMoviesTask.Wait();

        List<Result> movies = new();
        var tmdMovieInfo = tmdbMoviesTask.Result;

        if (tmdbMoviesTask.Result != null)
        {
            movies = tmdbMoviesTask.Result.results;
        }

        return movies;
    }

    [ChromelyRoute(Path = "/tmdbmoviescontroller/homepage")]
    public void HomePage(string movieId)
    {
        if (string.IsNullOrWhiteSpace(movieId))
        {
            return;
        }

        var tmdbMovieTask = Task.Run(() =>
        {
            return GetTmdbMovieAsync(movieId);
        });

        tmdbMovieTask.Wait();

        var movie = tmdbMovieTask.Result;
        if (movie != null && !string.IsNullOrWhiteSpace(movie.homepage))
        {
            BrowserLauncher.Open(_config.Platform, movie.homepage);
        }
    }

    private async Task<TmdMoviesInfo> GetTmdbMovieListAsync(string paramUrl)
    {
        var baseAddress = new Uri(TmdbBaseUrl);
        using var httpClient = new HttpClient { BaseAddress = baseAddress };
        httpClient.DefaultRequestHeaders.TryAddWithoutValidation("accept", "application/json");

        using var response = await httpClient.GetAsync(paramUrl);
        string responseData = await response.Content.ReadAsStringAsync();

        var options = new JsonSerializerOptions
        {
            ReadCommentHandling = JsonCommentHandling.Skip,
            AllowTrailingCommas = true
        };

        return JsonSerializer.Deserialize<TmdMoviesInfo>(responseData, options) ?? new TmdMoviesInfo();
    }

    private async Task<TmdMovie> GetTmdbMovieAsync(string movieId)
    {
        var baseAddress = new Uri(TmdbBaseUrl);
        using var httpClient = new HttpClient { BaseAddress = baseAddress };
        httpClient.DefaultRequestHeaders.TryAddWithoutValidation("accept", "application/json");

        using var response = await httpClient.GetAsync(TmdbGetMovieUrl(movieId));
        string responseData = await response.Content.ReadAsStringAsync();

        var options = new JsonSerializerOptions
        {
            ReadCommentHandling = JsonCommentHandling.Skip,
            AllowTrailingCommas = true
        };

        return JsonSerializer.Deserialize<TmdMovie>(responseData, options) ?? new TmdMovie();
    }
}