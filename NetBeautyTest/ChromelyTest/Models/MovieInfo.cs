namespace ChromelyTest.Models;

/// <summary>
/// The movie info.
/// </summary>
public class MovieInfo
{
    /// <summary>
    /// Initializes a new instance of the <see cref="MovieInfo"/> class.
    /// </summary>
    public MovieInfo()
    {
    }

    /// <summary>
    /// Initializes a new instance of the <see cref="MovieInfo"/> class.
    /// </summary>
    /// <param name="id">
    /// The id.
    /// </param>
    /// <param name="title">
    /// The title.
    /// </param>
    /// <param name="year">
    /// The year.
    /// </param>
    /// <param name="votes">
    /// The votes.
    /// </param>
    /// <param name="rating">
    /// The rating.
    /// </param>
    /// <param name="assembly">
    /// The assembly.
    /// </param>
    public MovieInfo(int id, string title, int year, int votes, double rating, string assembly)
    {
        Id = id;
        Title = title;
        Year = year;
        Votes = votes;
        Rating = rating;
        Date = DateTime.Now;
        RestfulAssembly = assembly;
    }

    public int Id { get; set; }

    public string? Title { get; set; }

    public int Year { get; set; }

    public int Votes { get; set; }

    public double Rating { get; set; }

    public DateTime Date { get; set; }

    public string? RestfulAssembly { get; set; }
}
