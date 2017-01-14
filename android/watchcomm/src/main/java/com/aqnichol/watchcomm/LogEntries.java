package com.aqnichol.watchcomm;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.HashMap;

/**
 * LogEntries is a JSON-compatible list of StatusHub log entries.
 */
public class LogEntries {
    public static class Entry {
        public String service;
        public String message;
    }

    private Entry[] entries;

    public LogEntries(Entry[] entries) {
        entries = entries;
    }

    public LogEntries(byte[] data) throws JSONException {
        ArrayList<Entry> res = new ArrayList<Entry>();
        JSONArray obj = new JSONArray(new String(data, StandardCharsets.UTF_8));
        for (int i = 0; i < obj.length(); i++) {
            JSONObject o = obj.getJSONObject(i);
            Entry e = new Entry();
            e.service = o.getString("serviceName");
            e.message = o.getString("message");
            res.add(e);
        }
        entries = res.toArray(new Entry[res.size()]);
    }

    public Entry[] getEntries() {
        return entries;
    }

    public byte[] marshal() {
        JSONObject[] e = new JSONObject[entries.length];
        for (int i = 0; i < entries.length; ++i) {
            Entry entry = entries[i];
            HashMap m = new HashMap();
            m.put("serviceName", entry.service);
            m.put("message", entry.message);
            e[i] = new JSONObject(m);
        }
        try {
            JSONArray a = new JSONArray(e);
            return a.toString().getBytes(StandardCharsets.UTF_8);
        } catch (JSONException e1) {
            return null;
        }
    }
}
