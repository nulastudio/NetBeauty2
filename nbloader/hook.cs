using System;
using System.Runtime.Loader;
using System.Reflection;

internal class StartupHook
{
    public static void Initialize()
    {
        /**
         * Run NBLoader in a isolated ALC
         * Avoid NBLoader's dependencies directly load into default ALC
         *
         * Fix https://github.com/nulastudio/NetBeauty2/issues/48
         * Fix https://github.com/nulastudio/NetBeauty2/issues/50
         */
        var alc = new AssemblyLoadContext("NBLoader");

        var asm = alc.LoadFromAssemblyPath(Assembly.GetExecutingAssembly().Location);

        var type = asm.GetType("NetBeauty.NBLoader");

        var register = type.GetMethod("RegisterALC", new Type[] { typeof(AssemblyLoadContext) });

        if (type != null && register != null)
        {
            var nbloader = Activator.CreateInstance(type);

            register.Invoke(nbloader, new object[] { AssemblyLoadContext.Default });
            register.Invoke(nbloader, new object[] { alc });
        }
    }
}
