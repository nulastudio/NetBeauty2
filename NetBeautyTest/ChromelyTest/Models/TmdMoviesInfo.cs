namespace ChromelyTest.Models;

public class Result
{
    public Result()
    {
        poster_path = string.Empty;
        backdrop_path = string.Empty;
        original_language = string.Empty;
        original_title = string.Empty;
        title = string.Empty;
        overview = string.Empty;
        _releaseDate = string.Empty;

        genre_ids = new List<int>();
    }
    public double popularity { get; set; }
    public int vote_count { get; set; }
    public bool video { get; set; }
    public string poster_path { get; set; }
    public int id { get; set; }
    public bool adult { get; set; }
    public string backdrop_path { get; set; }
    public string original_language { get; set; }
    public string original_title { get; set; }
    public List<int> genre_ids { get; set; }
    public string title { get; set; }
    public double vote_average { get; set; }
    public string overview { get; set; }

    private string _releaseDate;

    public string release_date
    {
        get
        {
            DateTime dateTime;
            if (DateTime.TryParse(_releaseDate, out dateTime))
            {
                return string.Format("{0:dddd, MMMM d, yyyy}", dateTime);
            }

            return _releaseDate;
        }
        set
        {
            _releaseDate = value;
        }
    }
}

public class TmdMoviesInfo
{
    public TmdMoviesInfo()
    {
        results = new List<Result>();
    }

    public int page { get; set; }
    public int total_results { get; set; }
    public int total_pages { get; set; }
    public List<Result> results { get; set; }
}

public class TmdMovie
{
    public TmdMovie()
    {
        homepage = string.Empty;
    }

    public int id { get; set; }
    public string homepage { get; set; }
}
