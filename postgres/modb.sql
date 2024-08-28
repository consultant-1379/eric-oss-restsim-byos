--
-- PostgreSQL database dump
--

-- Dumped from database version 11.18
-- Dumped by pg_dump version 11.18

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: modb; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.modb (
    uri character varying NOT NULL,
    data character varying
);


ALTER TABLE public.modb OWNER TO postgres;

--
-- Data for Name: modb; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.modb (uri, data) FROM stdin;
\.


--
-- Name: modb modb_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.modb
    ADD CONSTRAINT modb_pkey PRIMARY KEY (uri);


--
-- PostgreSQL database dump complete
--
