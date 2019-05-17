import scala.io.Source

val buildDescriptorMap = JSON
  .parseFull(Source.fromFile("sbtDescriptor.json").mkString)
  .get
  .asInstanceOf[Map[String, String]]

lazy val buildSettings = Seq(
  scalaVersion := "2.11.11",
)

lazy val root = (project in file("."))
  .settings(buildSettings)

libraryDependencies ++= Seq(
  jdbc,
  "org.scalatestplus.play" % "scalatestplus-play_2.11" % "2.0.0" % Test
)

dependencyOverrides += "com.fasterxml.jackson.core" % "jackson-databind" % "2.8.11.2"

resolvers ++= Seq(
  Resolver.url("Typesafe Ivy releases",
    url("https://repo.typesafe.com/typesafe/ivy-releases"))(Resolver.ivyStylePatterns)
)

// Play provides two styles of routers, one expects its actions to be injected, the
// other, legacy style, accesses its actions statically.
routesGenerator := InjectedRoutesGenerator

javaOptions in run ++= Seq(
  "-Xmx12G"
)

javaOptions in Universal ++= Seq(
  "-Dpidfile.path=/dev/null"
)

javaOptions in Test += "-Dconfig.file=conf/application.test.conf"

// Do not add API documentation into generated package
sources in (Compile, doc) := Seq.empty
publishArtifact in (Universal, packageBin) := true

// scala style
scalastyleConfig := baseDirectory.value / "scalastyle-production-config.xml"

// Whitesource
whitesourceProduct in ThisBuild               := "PRODUCT VERSION"
whitesourceOrgToken in ThisBuild              := "org-token"
whitesourceAggregateProjectName in ThisBuild  := "project-name"
whitesourceAggregateProjectToken in ThisBuild := "project-token"
whitesourceFailOnError in ThisBuild := false
