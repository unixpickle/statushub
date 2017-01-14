package com.aqnichol.statushub;

import android.content.SharedPreferences;

import com.aqnichol.watchcomm.CommException;
import com.aqnichol.watchcomm.LogEntries;
import com.aqnichol.watchcomm.Sender;
import com.google.android.gms.wearable.MessageEvent;
import com.google.android.gms.wearable.WearableListenerService;

import org.json.JSONException;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

public class RefreshService extends WearableListenerService {
    private static class LoginException extends Exception {
        LoginException(String msg) {
            super(msg);
        }
    }

    @Override
    public void onMessageReceived(MessageEvent messageEvent) {
        if (!messageEvent.getPath().equals("/refresh")) {
            return;
        }
        Sender c = new Sender(getApplicationContext());
        try {
            c.connect();
            String node = messageEvent.getSourceNodeId();
            try {
                LogEntries entries = fetchLogEntries();
                c.sendMessage(node, "/listing", entries.marshal());
            } catch (IOException e) {
                c.sendMessage(node, "/error", "fetch error: " + e.getMessage());
            } catch (JSONException e) {
                c.sendMessage(node, "/error", "JSON error: " + e.getMessage());
            } catch (LoginException e) {
                c.sendMessage(node, "/error", e.getMessage());
            }
        } catch (CommException e) {
        } finally {
            c.disconnect();
        }
    }

    private LogEntries fetchLogEntries() throws IOException, JSONException, LoginException {
        SharedPreferences prefs = getSharedPreferences("shhost", 0);
        String url = prefs.getString("rootURL", "");
        String password = prefs.getString("password", "");
        Client c = new Client(url);
        if (!c.login(password)) {
            throw new LoginException("login incorrect");
        }
        return c.overview();
    }
}
