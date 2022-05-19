using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Navigation;
using System.Windows.Shapes;

using WatsonWebsocket;

namespace WsServer
{
    /// <summary>
    /// Interaction logic for MainWindow.xaml
    /// </summary>
    public partial class MainWindow : Window
    {
        private WatsonWsServer server;
        private string lastClient;

        public MainWindow()
        {
            InitializeComponent();
        }

        private async void Start_Click(object sender, RoutedEventArgs e)
        {
            if (this.server != null && this.server.IsListening)
            {
                this.server.Stop();
            }

            var part = Address.Text.Split(":");
            var ip = part[0];
            var port = int.Parse(part[1]);

            WatsonWsServer server = new WatsonWsServer(ip, port, false);
            server.ClientConnected += ClientConnected;
            server.ClientDisconnected += ClientDisconnected;
            server.MessageReceived += MessageReceived;
            await server.StartAsync();

            if (server.IsListening)
            {
                this.server = server;
                Content.Text += $"Server Start\n\n";
            }
            else
            {
                Content.Text += $"Can not start the server\n\n";
            }
        }

        private void Stop_Click(object sender, RoutedEventArgs e)
        {
            if (this.server == null || !this.server.IsListening)
            {
                Content.Text += $"Server Unstart\n\n";
                return;
            }

            this.server.Stop();
            this.server.Dispose();
            this.server = null;

            Content.Text += $"Server Stop\n\n";
        }

        private async void Send_Click(object sender, RoutedEventArgs e)
        {
            var text = SendText.Text;

            if (!string.IsNullOrEmpty(text))
            {
                if (this.server == null || !this.server.IsListening)
                {
                    Content.Text += $"Server Unstart\n\n";
                }
                else
                {
                    if (string.IsNullOrEmpty(lastClient))
                    {
                        Content.Text += $"No Connected Client\n\n";
                    }
                    else
                    {
                        await this.server.SendAsync(lastClient, text);
                        SendText.Text = "";
                        Content.Text += $"Send:\n{text}\n\n";
                    }
                }
            }

            SendText.Focus();
        }

        private void MessageReceived(object sender, MessageReceivedEventArgs args)
        {
            Application.Current.Dispatcher.Invoke(() =>
            {
                Content.Text += $"Receive:\n{Encoding.UTF8.GetString(args.Data)}\n\n";
            });
        }

        private void ClientConnected(object sender, ClientConnectedEventArgs args)
        {
            lastClient = args.IpPort;
            Application.Current.Dispatcher.Invoke(() =>
            {
                Content.Text += $"Client Connect\n\n";
            });
        }

        private void ClientDisconnected(object sender, ClientDisconnectedEventArgs args)
        {
            Application.Current.Dispatcher.Invoke(() =>
            {
                Content.Text += $"Client Disconnect\n\n";
            });
        }
    }
}
