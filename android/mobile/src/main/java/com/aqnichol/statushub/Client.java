package com.aqnichol.statushub;

import android.text.TextUtils;
import android.util.Log;

import com.aqnichol.watchcomm.LogEntries;

import org.apache.commons.io.IOUtils;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.IOException;
import java.io.StringReader;
import java.io.StringWriter;
import java.net.CookieManager;
import java.net.HttpCookie;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.List;

/**
 * Communicate with a StatusHub server.
 */
public class Client {
    private static final int CONNECT_TIMEOUT = 10000;
    private static final int READ_TIMEOUT = 10000;

    private CookieManager cookieManager = new CookieManager();
    private String rootURL;

    Client(String url) {
        if (!url.endsWith("/")) {
            url += "/";
        }
        rootURL = url;
    }

    public boolean login(String password) throws IOException {
        HttpURLConnection conn = makeConnection(new URL(rootURL + "login"));
        try {
            conn.setDoOutput(true);
            conn.setDoInput(true);
            conn.setUseCaches(false);
            conn.setInstanceFollowRedirects(false);
            conn.setRequestMethod("POST");
            conn.setRequestProperty("Content-Type", "application/x-www-form-urlencoded");

            String postData = "password=" + URLEncoder.encode(password, "UTF-8");
            conn.setRequestProperty("Content-Length", "" + postData.getBytes().length);
            IOUtils.copy(new StringReader(postData), conn.getOutputStream(),
                    StandardCharsets.UTF_8);
            takeResponseCookies(conn);
            return conn.getHeaderField("Location").equals("/");
        } finally {
            conn.disconnect();
        }
    }

    public LogEntries overview() throws IOException, JSONException {
        HttpURLConnection conn = makeConnection(new URL(rootURL + "api/overview"));
        try {
            conn.connect();
            // TODO: don't convert to/from a String.
            StringWriter writer = new StringWriter();
            IOUtils.copy(conn.getInputStream(), writer, StandardCharsets.UTF_8);

            JSONObject obj = new JSONObject(writer.toString());
            String dataObj = obj.getJSONArray("data").toString();

            return new LogEntries(dataObj.getBytes(StandardCharsets.UTF_8));
        } finally {
            conn.disconnect();
        }
    }

    private HttpURLConnection makeConnection(URL url) throws IOException {
        HttpURLConnection connection = (HttpURLConnection)url.openConnection();
        connection.setConnectTimeout(CONNECT_TIMEOUT);
        connection.setReadTimeout(READ_TIMEOUT);
        putRequestCookies(connection);
        return connection;
    }

    private void takeResponseCookies(HttpURLConnection connection) {
        List<String> cookiesHeader = connection.getHeaderFields().get("Set-Cookie");
        if (cookiesHeader != null) {
            for (String cookie : cookiesHeader) {
                for (HttpCookie parsedCookie : HttpCookie.parse(cookie))
                    cookieManager.getCookieStore().add(null, parsedCookie);
            }
        }
    }

    private void putRequestCookies(HttpURLConnection connection) {
        List<HttpCookie> cookies = cookieManager.getCookieStore().getCookies();
        if (cookies.size() > 0) {
            connection.setRequestProperty("Cookie", TextUtils.join(";", cookies));
        }
    }
}
