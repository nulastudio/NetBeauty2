using System;
using System.Xml;
using System.Xml.Serialization;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace Web.Pages;

public class PrivacyModel : PageModel
{
    private readonly ILogger<PrivacyModel> _logger;

    public int Id { get; private set; }

    public PrivacyModel(ILogger<PrivacyModel> logger)
    {
        _logger = logger;
    }

    public void OnGet(int bid, int cid)
    {
        this.Id = bid;

        //var entity = new UrlRewriteConfig
        //{
        //    Items = new UrlRule[] { new UrlRule
        //    {
        //        Name="Test1",
        //        Url="id=3&p=4",
        //        Rewrite="home/3/4"
        //    },
        //    new UrlRule
        //    {
        //        Name="Test1",
        //        Url="id=3&p=4",
        //        Redirect="home/3/4"
        //    }
        //    }
        //};


        //XmlSerializer serializer = new XmlSerializer(typeof(UrlRewriteConfig));
        //using (StringWriter textWriter = new StringWriter())
        //{
        //    serializer.Serialize(textWriter, entity);
        //    var xml = textWriter.ToString();
        //}

        var xmlString = @"<?xml version=""1.0"" encoding=""utf-8""?>
<xml>
  <UrlRules>
    <UrlRule Name=""Test1"">
      <Url><![CDATA[id=3&p=4]]></Url>
      <Rewrite>home/3/4</Rewrite>
    </UrlRule>
    <UrlRule Name=""Test1"">
      <Url>id=3&amp;p=4</Url>
      <Redirect>home/3/4</Redirect>
    </UrlRule>
  </UrlRules>
</xml>"
        ;

        var serializer = new XmlSerializer(typeof(UrlRewriteConfig));
        using (StringReader reader = new StringReader(xmlString))
        {
            var person = (UrlRewriteConfig)serializer.Deserialize(reader);
            // 现在可以使用person对象中的Name和Age属性
        }

    }



}

[XmlRootAttribute("xml")]
public class UrlRewriteConfig
{
    [XmlArrayAttribute("UrlRules")]
    public UrlRule[] Items;
}

public class UrlRule
{

    [XmlAttribute("Name")]
    public string Name { get; set; }
    public string Url { get; set; }
    public string Rewrite { get; set; }
    public string Redirect { get; set; }


}





