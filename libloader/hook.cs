using System;
using System.Runtime.Loader;
using System.Reflection;

internal class StartupHook
{
    public static void Initialize()
    {
        /**
         * Run Loader in a isolated ALC
         * Avoid Loader's dependencies directly load into default ALC
         *
         * Fix https://github.com/nulastudio/NetBeauty2/issues/48
         * Fix https://github.com/nulastudio/NetBeauty2/issues/50
         */
        var alc = new AssemblyLoadContext("LibLoader");

        var asm = alc.LoadFromAssemblyPath(Assembly.GetExecutingAssembly().Location);

        var type = asm.GetType("NetBeauty.LibLoader");

        var register = type.GetMethod("RegisterALC", new Type[] { typeof(AssemblyLoadContext) });

        if (type != null && register != null)
        {
            var loader = Activator.CreateInstance(type);

            register.Invoke(loader, new object[] { AssemblyLoadContext.Default });
            register.Invoke(loader, new object[] { alc });
        }
    }
}
